package get

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"gomarketplace_api/internal/wholesaler/pkg/clients"
	"gomarketplace_api/internal/wildberries/business/dto/responses"
	request2 "gomarketplace_api/internal/wildberries/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/business/services"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	WorkerCount    = 5
	MaxRetries     = 3
	RetryInterval  = 2 * time.Second
	RequestTimeout = 100 * time.Second
)

var (
	ErrNoMoreData         = errors.New("no more data to process")
	ErrRateLimiter        = errors.New("rate limiter error")
	ErrConnectionAborted  = errors.New("connection aborted")
	ErrTotalLimitReached  = errors.New("total limit reached")
	ErrContextCanceled    = errors.New("context canceled")
	ErrFailedAfterRetries = errors.New("failed after retries")
)

type Config struct {
	WorkerCount    int
	MaxRetries     int
	RetryInterval  time.Duration
	RequestTimeout time.Duration
}

type SearchEngine struct {
	db      *sql.DB
	auth    services.AuthEngine
	writer  io.Writer
	config  Config
	limiter *rate.Limiter
}

func NewSearchEngine(db *sql.DB, auth services.AuthEngine, writer io.Writer, config Config) *SearchEngine {
	return &SearchEngine{
		db:      db,
		auth:    auth,
		writer:  writer,
		config:  config,
		limiter: rate.NewLimiter(rate.Every(time.Minute/70), 10),
	}
}

const postNomenclature = "https://content-api.wildberries.ru/content/v2/get/cards/list"

func (d *SearchEngine) GetNomenclatures(settings request2.Settings, locale string) (*responses.NomenclatureResponse, error) {
	url := postNomenclature
	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: d.config.RequestTimeout}

	requestBody, err := settings.CreateRequestBody()
	if err != nil {
		return nil, fmt.Errorf("creating request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	d.auth.SetApiKey(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var nomenclatureResponse responses.NomenclatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&nomenclatureResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &nomenclatureResponse, nil
}

type SafeCursorManager struct {
	mu          sync.Mutex
	usedCursors map[string]bool
	lastCursor  request2.Cursor
}

func NewSafeCursorManager() *SafeCursorManager {
	return &SafeCursorManager{
		usedCursors: make(map[string]bool),
	}
}

func (scm *SafeCursorManager) GetUniqueCursor(nmID int, updatedAt string) (request2.Cursor, bool) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	cursorKey := fmt.Sprintf("%d_%s", nmID, updatedAt)

	if scm.usedCursors[cursorKey] {
		return request2.Cursor{}, false
	}

	scm.usedCursors[cursorKey] = true
	cursor := request2.Cursor{
		NmID:      nmID,
		UpdatedAt: updatedAt,
		Limit:     100,
	}
	scm.lastCursor = cursor

	return cursor, true
}

func (d *SearchEngine) GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(
	ctx context.Context,
	settings request2.Settings,
	locale string,
	nomenclatureChan chan response.Nomenclature,
) error {
	defer log.Printf("Search engine finished its job.")

	limit := settings.Cursor.Limit
	log.Printf("Getting Wildberries nomenclatures with limit: %d", limit)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*30)
	defer cancel()

	cursorManager := NewSafeCursorManager()
	taskChan := make(chan request2.Cursor, d.config.WorkerCount)
	errChan := make(chan error, d.config.WorkerCount)

	var (
		wg             sync.WaitGroup
		totalProcessed atomic.Int32
	)

	for i := 0; i < d.config.WorkerCount; i++ {
		wg.Add(1)
		go d.safeWorker(
			ctx,
			i,
			&wg,
			taskChan,
			nomenclatureChan,
			errChan,
			settings,
			locale,
			limit,
			&totalProcessed,
			cursorManager,
		)
	}

	log.Printf("Sending initial task to taskChan")
	select {
	case taskChan <- request2.Cursor{NmID: 0, UpdatedAt: "", Limit: 100}:
		log.Printf("Initial task sent successfully")
	case <-ctx.Done():
		log.Printf("Context cancelled while sending initial task")
		return ctx.Err()
	}

	go func() {
		wg.Wait()
		log.Printf("Starting cleanup goroutine")
		log.Printf("All workers finished, closing channels")
		close(taskChan)
		close(errChan)
		log.Printf("Channels closed")
	}()

	for err := range errChan {
		if err == nil {
			continue
		}
		log.Printf("Received error from worker: %v", err)

		if errors.Is(err, ErrNoMoreData) {
			log.Printf("No more data to process")
			continue
		}
		if errors.Is(err, ErrTotalLimitReached) {
			log.Printf("Total limit reached")
			continue
		}
		if errors.Is(err, ErrContextCanceled) {
			log.Printf("Context was canceled")
			continue
		}

		cancel()
		return fmt.Errorf("worker error: %w", err)
	}

	return nil
}

func (d *SearchEngine) safeWorker(
	ctx context.Context,
	workerID int,
	wg *sync.WaitGroup,
	taskChan chan request2.Cursor,
	nomenclatureChan chan<- response.Nomenclature,
	errChan chan error,
	settings request2.Settings,
	locale string,
	totalLimit int,
	totalProcessed *atomic.Int32,
	cursorManager *SafeCursorManager,
) {
	defer func() {
		log.Printf("Worker %d: job is done.", workerID)
		wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			errChan <- ErrContextCanceled
			return
		case cursor, ok := <-taskChan:
			if !ok {
				return
			}

			uniqueCursor, ok := cursorManager.GetUniqueCursor(cursor.NmID, cursor.UpdatedAt)
			if !ok {
				continue
			}

			err := d.processSafeCursorTask(
				ctx,
				workerID,
				uniqueCursor,
				nomenclatureChan,
				settings,
				locale,
				totalLimit,
				totalProcessed,
				taskChan,
			)

			if err != nil && !errors.Is(err, ErrNoMoreData) && !errors.Is(err, ErrTotalLimitReached) {
				errChan <- err
			}
		}
	}
}

func (d *SearchEngine) processSafeCursorTask(
	ctx context.Context,
	workerID int,
	cursor request2.Cursor,
	nomenclatureChan chan<- response.Nomenclature,
	settings request2.Settings,
	locale string,
	totalLimit int,
	totalProcessed *atomic.Int32,
	taskChan chan<- request2.Cursor,
) error {
	if err := d.limiter.Wait(ctx); err != nil {
		log.Printf("Worker %d: Rate limiter error: %v", workerID, err)
		return fmt.Errorf("%w: %v", ErrRateLimiter, err)
	}

	settings.Cursor = cursor
	log.Printf("Worker %d: Fetching nomenclatures with cursor: NmID=%d, UpdatedAt=%v, Limit=%d",
		workerID, cursor.NmID, cursor.UpdatedAt, cursor.Limit)

	nomenclatureResponse, err := d.retryGetNomenclatures(ctx, settings, &cursor, locale)
	if err != nil {
		if errors.Is(err, ErrContextCanceled) {
			return err
		}
		return fmt.Errorf("failed to get nomenclatures: %w", err)
	}

	lastNomenclature := nomenclatureResponse.Data[len(nomenclatureResponse.Data)-1]

	if len(nomenclatureResponse.Data) == cursor.Limit {
		select {
		case taskChan <- request2.Cursor{
			NmID:      lastNomenclature.NmID,
			UpdatedAt: lastNomenclature.UpdatedAt,
			Limit:     cursor.Limit,
		}:
		case <-ctx.Done():
			return ErrContextCanceled
		}
	}

	for _, nomenclature := range nomenclatureResponse.Data {
		totalProcessed.Add(1)
		if totalProcessed.Load() >= int32(totalLimit) {
			return ErrTotalLimitReached
		}

		select {
		case <-ctx.Done():
			return ErrContextCanceled
		case nomenclatureChan <- nomenclature:
		}
	}

	if len(nomenclatureResponse.Data) < cursor.Limit {
		defer close(nomenclatureChan)
		return ErrNoMoreData
	}

	return nil
}

func (d *SearchEngine) retryGetNomenclatures(
	ctx context.Context,
	settings request2.Settings,
	cursor *request2.Cursor,
	locale string,
) (*responses.NomenclatureResponse, error) {
	var lastErr error

	for retry := 0; retry < d.config.MaxRetries; retry++ {
		select {
		case <-ctx.Done():
			return nil, ErrContextCanceled
		default:
		}

		settings.Cursor = *cursor
		nomenclatureResponse, err := d.GetNomenclatures(settings, locale)
		if err == nil {
			return nomenclatureResponse, nil
		}

		if errors.Is(err, ErrConnectionAborted) {
			log.Printf("Retrying to get nomenclatures due to connection error. Attempt: %d", retry+1)
			lastErr = err
			time.Sleep(d.config.RetryInterval)
			continue
		}

		return nil, fmt.Errorf("failed to get nomenclatures: %w", err)
	}

	return nil, fmt.Errorf("%w: %v", ErrFailedAfterRetries, lastErr)
}

func (d *SearchEngine) prepareGlobalIDs() (map[int]struct{}, error) {
	client := clients.NewGlobalIDsClient("http://localhost:8081", d.writer)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.Fetch(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching Global IDs: %w", err)
	}

	globalIDs, ok := result.([]int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for global IDs: expected []int, got %T", result)
	}

	globalIDsMap := make(map[int]struct{}, len(globalIDs))
	for _, globalID := range globalIDs {
		globalIDsMap[globalID] = struct{}{}
	}
	return globalIDsMap, nil
}

/*
Возвращает число обновленных(добавленных) карточек
*/
func (d *SearchEngine) UploadToDb(settings request2.Settings, locale string) (int, error) {
	log.Printf("Updating wildberries.nomenclatures")
	log.SetPrefix("NM UPDATER | ")

	updated := 0
	wg := sync.WaitGroup{}
	var mu sync.Mutex
	// Получаем существующие номенклатуры из БД
	existIDs, err := d.GetDBNomenclatures()
	if err != nil {
		return updated, err
	}
	client := clients.NewGlobalIDsClient("http://localhost:8081", d.writer)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.Fetch(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf("error fetching Global IDs: %w", err)
	}

	globalIDs, ok := result.([]int)
	if !ok {
		return -1, fmt.Errorf("unexpected type for global IDs: expected []int, got %T", result)
	}

	globalIDsFromDBMap := make(map[int]struct{}, len(globalIDs))
	for _, globalID := range globalIDs {
		globalIDsFromDBMap[globalID] = struct{}{}
	}

	nomenclatureChan := make(chan response.Nomenclature)
	ctx, cancel = context.WithCancel(context.Background())
	log.Println("Fetching and sending nomenclatures to the channel...")

	defer cancel()
	go func() {
		if err := d.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(ctx, settings, locale, nomenclatureChan); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
	}()

	var uploadPacket []interface{}
	loadingChan := make(chan []interface{})
	saw := 0
	errs := make(map[int]string)

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			log.Printf("Nomenclature upload channel closed. Returning")
		}()
		for nomenclature := range nomenclatureChan {
			mu.Lock()
			saw++
			mu.Unlock()
			id, err := nomenclature.GlobalID()
			if err != nil {
				mu.Lock()
				errs[id] = fmt.Sprintf("ID: %d -- Nomenclature upload failed: %s", nomenclature.VendorCode, err)
				mu.Unlock()
				continue
			}
			if _, ok := globalIDsFromDBMap[id]; !ok {
				mu.Lock()
				errs[id] = fmt.Sprintf("ID: %d -- GlobalIDMap not contains this id: %s", nomenclature.VendorCode, err)
				mu.Unlock()
				continue
			}
			if _, ok := existIDs[id]; ok {
				continue
			}

			mu.Lock()
			uploadPacket = append(uploadPacket, id, nomenclature.NmID, nomenclature.ImtID,
				nomenclature.NmUUID, nomenclature.VendorCode, nomenclature.SubjectID,
				nomenclature.Brand, nomenclature.CreatedAt, nomenclature.UpdatedAt)
			mu.Unlock()

			if len(uploadPacket)/9 == 100 {
				loadingChan <- uploadPacket
				mu.Lock()
				uploadPacket = []interface{}{}
				mu.Unlock()
			}
		}
	}()

	go func() {
		defer func() {
			log.Printf("Nomenclature load to db channel closed. Returning")
		}()
		for loading := range loadingChan {
			err = d.insertBatchNomenclatures(loading)
			if err != nil {
				log.Printf("Error during upload nomenclatures in db")
			}
			log.Printf("Successfully updates nomenclatures in db")
		}
	}()

	wg.Wait()

	if len(uploadPacket) > 0 {
		loadingChan <- uploadPacket
		uploadPacket = []interface{}{}
	}
	close(loadingChan)

	log.Printf("It looks like all the data is up to date\nSaw: %d", saw)

	for k, v := range errs {
		log.Printf("Error uploading nomenclature: %s. Details: %v", k, v)
	}

	log.SetPrefix("")
	return updated, nil
}

func (d *SearchEngine) CheckTotalNmCount(settings request2.Settings, locale string) (int, error) {
	log.Printf("You are testing search engine")
	log.SetPrefix("[ TEST ] ")

	var count int
	var msc int

	if _, err := d.prepareGlobalIDs(); err != nil {
		return 0, fmt.Errorf("error")
	}

	nomenclatureChan := make(chan response.Nomenclature)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go func() {
		if err := d.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(ctx, settings, locale, nomenclatureChan); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
	}()

	for nm := range nomenclatureChan {
		if _, err := nm.GlobalID(); err != nil {
			msc++
		}
		count++
	}

	log.Printf("found : %d msc articulars", msc)

	return count, nil
}

func (d *SearchEngine) insertBatchNomenclatures(batch []interface{}) error {
	// Максимальное количество записей в одной пачке
	const maxBatchSize = 900 // 900 записей = 8100 параметров (по 9 параметров на запись)
	batch = removeDuplicateRecords(batch)
	// Разделение на подмассивы
	for start := 0; start < len(batch); start += maxBatchSize * 9 {
		end := start + maxBatchSize*9
		if end > len(batch) {
			end = len(batch) // Учитываем остаток
		}

		// Текущая пачка данных
		currentBatch := batch[start:end]
		numRecords := len(currentBatch) / 9 // Число записей в текущей пачке

		// Генерация частей запроса для текущей пачки
		valueStrings := []string{}
		for i := 0; i < numRecords; i++ {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
		}

		// Полный SQL-запрос
		query := fmt.Sprintf(`
			INSERT INTO wildberries.nomenclatures (global_id, nm_id, imt_id, nm_uuid, vendor_code, subject_id, wb_brand, created_at, updated_at)
			VALUES 
				%s
			ON CONFLICT (global_id) DO UPDATE
			SET 
				nm_id = EXCLUDED.nm_id,
				imt_id = EXCLUDED.imt_id,
				nm_uuid = EXCLUDED.nm_uuid,
				vendor_code = EXCLUDED.vendor_code,
				subject_id = EXCLUDED.subject_id,
				wb_brand = EXCLUDED.wb_brand,
				created_at = LEAST(nomenclatures.created_at, EXCLUDED.created_at),
				updated_at = GREATEST(nomenclatures.updated_at, EXCLUDED.updated_at);
		`, strings.Join(valueStrings, ", "))

		// Выполняем запрос с текущей пачкой параметров
		if _, err := d.db.Exec(query, currentBatch...); err != nil {
			return err
		}
	}

	return nil
}

func removeDuplicateRecords(batch []interface{}) []interface{} {
	unique := make(map[interface{}]bool) // Хранилище уникальных записей
	result := []interface{}{}            // Результирующий срез

	// Итерируемся по данным в батче
	for i := 0; i < len(batch); i += 9 { // Группируем записи по 9 параметров
		recordKey := fmt.Sprintf("%v", batch[i]) // Используем `global_id` как уникальный ключ
		if _, exists := unique[recordKey]; !exists {
			unique[recordKey] = true
			result = append(result, batch[i:i+9]...) // Добавляем запись в результирующий срез
		}
	}

	return result
}

func (d *SearchEngine) GetDBNomenclatures() (map[int]struct{}, error) {
	// запрос для получения списка category_id
	query := `SELECT global_id FROM wildberries.nomenclatures`

	// создаем срез для хранения category_id
	nmIDs := make(map[int]struct{}, 1)
	rows, err := d.db.Query(query)
	if err != nil {
		return map[int]struct{}{}, fmt.Errorf("ошибка выполнения запроса для категорий: %w", err)
	}
	defer rows.Close()

	// заполняем срез category_id из результата запроса
	for rows.Next() {
		var catID int
		if err := rows.Scan(&catID); err != nil {
			return map[int]struct{}{}, fmt.Errorf("ошибка сканирования cat_id: %w", err)
		}
		nmIDs[catID] = struct{}{}
	}

	// проверяем ошибки после цикла rows.Next()
	if err := rows.Err(); err != nil {
		return map[int]struct{}{}, fmt.Errorf("ошибка чтения строк: %w", err)
	}

	return nmIDs, nil
}

// ProcessNomenclatures запускает worker'ы для обработки номенклатур из канала.
// processFunc обрабатывает отдельную номенклатуру и возвращает (optional) модель для загрузки.
func ProcessNomenclatures(
	nomenclatureChan <-chan response.Nomenclature,
	workers int,
	processedItems *sync.Map,
	processFunc func(workerID int, nomenclature response.Nomenclature) (request2.Model, bool),
	outputChan chan<- request2.Model,
) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for nomenclature := range nomenclatureChan {
				// Проверка на дубликаты по VendorCode
				if _, loaded := processedItems.LoadOrStore(nomenclature.VendorCode, true); loaded {
					continue
				}
				if model, ok := processFunc(workerID, nomenclature); ok {
					outputChan <- model
				}
			}
		}(i)
	}
	wg.Wait()
	close(outputChan)
}

// UploadWorker отправляет модели на сервер с учетом rate limiter'а.
func UploadWorker(
	ctx context.Context,
	uploadChan <-chan request2.Model,
	uploadURL string,
	limiter *rate.Limiter,
	processAndUploadFunc func(string, request2.Model) (int, error),
	updatedCount *atomic.Int32,
) {
	for model := range uploadChan {
		if err := limiter.Wait(ctx); err != nil {
			log.Printf("Limiter error: %s", err)
			continue
		}
		count, err := processAndUploadFunc(uploadURL, model)
		if err != nil {
			log.Printf("Error during upload: %s", err)
			continue
		}
		updatedCount.Add(int32(count))
	}
}
