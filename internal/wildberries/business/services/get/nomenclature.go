package get

import (
	"context"
	"database/sql"
	"encoding/json"
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
	WorkerCount = 5
)

// SearchEngine -- сервис по работе с номенклатурами. get-update
type SearchEngine struct {
	db *sql.DB
	services.AuthEngine
	writer io.Writer
}

func NewSearchEngine(db *sql.DB, auth services.AuthEngine, writer io.Writer) *SearchEngine {
	return &SearchEngine{db: db, AuthEngine: auth, writer: writer}
}

const postNomenclature = "https://content-api.wildberries.ru/content/v2/get/cards/list"

func (d *SearchEngine) GetNomenclatures(settings request2.Settings, locale string) (*responses.NomenclatureResponse, error) {
	url := postNomenclature
	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: 100 * time.Second}

	requestBody, err := settings.CreateRequestBody()
	if err != nil {
		return nil, fmt.Errorf("creating request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return nil, err
	}

	d.SetApiKey(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var nomenclatureResponse responses.NomenclatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&nomenclatureResponse); err != nil {
		return nil, err
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
	nomenclatureChan chan<- response.Nomenclature,
	responseLimiter *rate.Limiter,
) error {
	defer func() {
		log.Printf("Search engine finished its job.")
	}()

	limit := settings.Cursor.Limit
	log.Printf("Getting Wildberries nomenclatures with limit: %d", limit)

	globalIDsMap, err := d.prepareGlobalIDs()
	if err != nil {
		return fmt.Errorf("failed to prepare global IDs: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// создаем потокобезопасный менеджер курсоров
	cursorManager := NewSafeCursorManager()

	packetSizes := divideLimitsToPackets(limit, 100)
	taskChan := make(chan request2.Cursor, len(packetSizes))
	errChan := make(chan error, WorkerCount)
	done := make(chan struct{})

	var (
		wg             sync.WaitGroup
		totalProcessed atomic.Int32
	)

	// запускаем воркеров
	for i := 0; i < WorkerCount; i++ {
		wg.Add(1)
		go d.safeWorker(
			ctx,
			i,
			&wg,
			taskChan,
			nomenclatureChan,
			errChan,
			responseLimiter,
			settings,
			locale,
			limit,
			&totalProcessed,
			globalIDsMap,
			cursorManager,
		)
	}

	// начальный курсор
	initialCursor := request2.Cursor{NmID: 0, UpdatedAt: "", Limit: 100}
	taskChan <- initialCursor

	// распределение задач
	go func() {
		defer close(taskChan)
		for _, size := range packetSizes {
			select {
			case <-ctx.Done():
				return
			case cursor := <-taskChan:
				cursor.Limit = size
				taskChan <- cursor
			}
		}
	}()

	// ожидание завершения
	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			if !strings.Contains(err.Error(), "OK. worker") {
				cancel()
				close(nomenclatureChan)
				return err
			}
			return nil
		case <-done:
			close(nomenclatureChan)
			return nil
		}
	}
}

func (d *SearchEngine) safeWorker(
	ctx context.Context,
	workerID int,
	wg *sync.WaitGroup,
	taskChan chan request2.Cursor,
	nomenclatureChan chan<- response.Nomenclature,
	errChan chan<- error,
	responseLimiter *rate.Limiter,
	settings request2.Settings,
	locale string,
	totalLimit int,
	totalProcessed *atomic.Int32,
	globalIDsMap map[int]struct{},
	cursorManager *SafeCursorManager,
) {
	defer func() {
		log.Println(fmt.Sprintf("Worker's %d job is done.", workerID))
		wg.Done()
	}()

	for cursor := range taskChan {
		uniqueCursor, ok := cursorManager.GetUniqueCursor(cursor.NmID, cursor.UpdatedAt)
		if !ok {
			continue
		}

		if err := d.processSafeCursorTask(
			ctx,
			workerID,
			uniqueCursor,
			nomenclatureChan,
			responseLimiter,
			settings,
			locale,
			totalLimit,
			totalProcessed,
			taskChan,
			globalIDsMap,
		); err != nil {
			select {
			case errChan <- err:
			default:
			}
			return
		}
	}
}

func (d *SearchEngine) processSafeCursorTask(
	ctx context.Context,
	workerID int,
	cursor request2.Cursor,
	nomenclatureChan chan<- response.Nomenclature,
	responseLimiter *rate.Limiter,
	settings request2.Settings,
	locale string,
	totalLimit int,
	totalProcessed *atomic.Int32,
	taskChan chan<- request2.Cursor,
	globalIDsMap map[int]struct{},
) error {
	if err := responseLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	settings.Cursor = cursor
	log.Printf("Worker %d: Fetching nomenclatures with cursor: NmID=%d, UpdatedAt=%v, Limit=%d",
		workerID, cursor.NmID, cursor.UpdatedAt, cursor.Limit)

	nomenclatureResponse, err := d.retryGetNomenclatures(ctx, settings, &cursor, locale)
	if err != nil {
		return fmt.Errorf("failed to get nomenclatures: %w", err)
	}

	if len(nomenclatureResponse.Data) == 0 {
		return fmt.Errorf("OK. worker %d: No more data to process", workerID)
	}

	lastNomenclature := nomenclatureResponse.Data[len(nomenclatureResponse.Data)-1]

	// отправляем следующий курсор в канал задач
	select {
	case taskChan <- request2.Cursor{
		NmID:      lastNomenclature.NmID,
		UpdatedAt: lastNomenclature.UpdatedAt,
	}:
	default:
	}

	for _, nomenclature := range nomenclatureResponse.Data {
		nmGlobalID, err := nomenclature.GlobalID()
		if err != nil {
			continue
		}
		if _, exists := globalIDsMap[nmGlobalID]; !exists {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case nomenclatureChan <- nomenclature:
			if totalProcessed.Add(1) >= int32(totalLimit) {
				log.Printf("Total processed items have reached the limit (%d)", totalLimit)
				return nil
			}
		}
	}

	return nil
}

func (d *SearchEngine) retryGetNomenclatures(
	ctx context.Context,
	settings request2.Settings,
	cursor *request2.Cursor,
	locale string,
) (*responses.NomenclatureResponse, error) {
	const (
		maxRetries    = 3
		retryInterval = 2 * time.Second
	)

	var lastErr error
	for retry := 0; retry < maxRetries; retry++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		settings.Cursor = *cursor
		nomenclatureResponse, err := d.GetNomenclatures(settings, locale)
		if err == nil {
			return nomenclatureResponse, nil
		}

		if !strings.Contains(err.Error(), "wsarecv: An established connection was aborted") {
			return nil, err
		}

		lastErr = err
		log.Printf("Retrying to get nomenclatures due to connection error. Attempt: %d", retry+1)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (d *SearchEngine) prepareGlobalIDs() (map[int]struct{}, error) {
	client := clients.NewGlobalIDsClient("http://localhost:8081", d.writer)
	globalIDs, err := client.FetchGlobalIDs()
	if err != nil {
		return nil, fmt.Errorf("error fetching Global IDs: %w", err)
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
	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/50), 50)
	wg := sync.WaitGroup{}
	var mu sync.Mutex
	// Получаем существующие номенклатуры из БД
	existIDs, err := d.GetDBNomenclatures()
	if err != nil {
		return updated, err
	}
	client := clients.NewGlobalIDsClient("http://localhost:8081", d.writer)

	// Инициализируем мапу globalIDsFromDBMap
	globalIDsFromDB, err := client.FetchGlobalIDs()
	if err != nil {
		log.Fatalf("Error fetching Global IDs: %s", err)
	}
	globalIDsFromDBMap := make(map[int]struct{}, len(globalIDsFromDB))
	for _, globalID := range globalIDsFromDB {
		globalIDsFromDBMap[globalID] = struct{}{}
	}

	nomenclatureChan := make(chan response.Nomenclature)
	ctx, cancel := context.WithCancel(context.Background())
	log.Println("Fetching and sending nomenclatures to the channel...")

	defer cancel()
	go func() {
		if err := d.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(ctx, settings, locale, nomenclatureChan, responseLimiter); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
	}()

	var uploadPacket []interface{}
	loadingChan := make(chan []interface{})
	saw := 0
	errors := make(map[int]string)

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
				errors[id] = fmt.Sprintf("ID: %d -- Nomenclature upload failed: %s", nomenclature.VendorCode, err)
				mu.Unlock()
				continue
			}
			if _, ok := globalIDsFromDBMap[id]; !ok {
				mu.Lock()
				errors[id] = fmt.Sprintf("ID: %d -- GlobalIDMap not contains this id: %s", nomenclature.VendorCode, err)
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

	for k, v := range errors {
		log.Printf("Error uploading nomenclature: %s. Details: %v", k, v)
	}

	log.SetPrefix("")
	return updated, nil
}

func (d *SearchEngine) CheckTotalNmCount(settings request2.Settings, locale string) (int, error) {
	log.Printf("You are testing search engine")
	log.SetPrefix("TEST | ")

	var count atomic.Int32
	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/50), 50)

	nomenclatureChan := make(chan response.Nomenclature)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go func() {
		if err := d.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(ctx, settings, locale, nomenclatureChan, responseLimiter); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
	}()

	for range nomenclatureChan {
		count.Add(1)
	}

	return int(count.Load()), nil
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

func contains(globalIDsMap map[int]struct{}, globalId int) bool {
	_, exists := globalIDsMap[globalId]
	return exists
}

func divideLimitsToPackets(totalCount int, packetSize int) []int {
	var packets []int
	for count := totalCount; count > 0; count -= packetSize {
		if count >= packetSize {
			packets = append(packets, packetSize)
		} else {
			packets = append(packets, count)
		}
	}
	return packets
}
