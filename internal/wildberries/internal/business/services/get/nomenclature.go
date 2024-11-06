package get

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"gomarketplace_api/internal/wholesaler/pkg/clients"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const HtmlRequestLimit = 3

// NomenclatureService -- сервис по работе с номенклатурами. get-update
type NomenclatureService struct {
	db            *sql.DB
	updateService get.UpdateService
}

func NewNomenclatureUpdateGetter(db *sql.DB, updateService get.UpdateService) *NomenclatureService {
	return &NomenclatureService{db: db, updateService: updateService}
}

const postNomenclature = "https://content-api.wildberries.ru/content/v2/get/cards/list"

func (d *NomenclatureService) GetNomenclatures(settings request.Settings, locale string) (*responses.NomenclatureResponse, error) {
	url := postNomenclature
	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: 20 * time.Second}

	requestBody, err := settings.CreateRequestBody()
	if err != nil {
		return nil, fmt.Errorf("creating request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return nil, err
	}

	services.SetAuthorizationHeader(req)
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

func (d *NomenclatureService) GetNomenclaturesWithLimitConcurrency(limit int, locale string) (map[int]response.Nomenclature, error) {
	log.Printf("Getting wildberries nomenclatures with limit: %d", limit)

	data := sync.Map{}
	limiter := rate.NewLimiter(rate.Limit(HtmlRequestLimit), HtmlRequestLimit)
	client := clients.NewGlobalIDsClient("http://localhost:8081")

	globalIDs, err := client.FetchGlobalIDs()
	if err != nil {
		log.Fatalf("Error fetching Global IDs: %s", err)
	}
	globalIDsMap := make(map[int]struct{}, len(globalIDs))
	for _, globalID := range globalIDs {
		globalIDsMap[globalID] = struct{}{}
	}

	cursor := request.Cursor{Limit: limit}
	sort := request.Sort{}
	filter := request.Filter{WithPhoto: -1}
	packetSizes := divideLimitsToPackets(limit, 100)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var activeGoroutines int
	limitChan := make(chan int, len(packetSizes))
	dataChan := make(chan response.Nomenclature)
	doneOnce := sync.Once{}
	totalProcessed := 0 // Переменная для отслеживания общего количества добавленных элементов
	totalErrors := 0
	log.Printf("Packet sizes: %v", packetSizes)

	// Горутина для сбора данных из dataChan в data map
	wg.Add(1)
	go func() {
		defer wg.Done()
		for nomenclature := range dataChan {
			globalId, err := nomenclature.GlobalID()
			if err != nil || globalId == 0 {
				log.Printf("[!] Not SPB articular (%s)", nomenclature.VendorCode)
				totalErrors++
				continue
			}
			data.LoadOrStore(globalId, nomenclature)
		}
	}()

	for i := 0; i < len(packetSizes); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mu.Lock()
			activeGoroutines++
			mu.Unlock()
			defer func() {
				mu.Lock()
				activeGoroutines--
				if activeGoroutines == 0 {
					log.Printf("It seems that we got all items.")
					doneOnce.Do(func() { close(dataChan) })
				}
				mu.Unlock()
			}()

			for {
				select {
				case packetLimit, ok := <-limitChan:
					if !ok {
						log.Printf("Goroutine %d: limitChan closed", i)
						return
					}

					if err := limiter.Wait(context.Background()); err != nil {
						log.Printf("Rate limiter error: %s", err)
						return
					}

					for {
						mu.Lock()
						cursor.Limit = packetLimit
						log.Printf("Goroutine %d: Fetching nomenclatures with packetLimit %d", i, packetLimit)
						log.Printf("Cursor state before request: NmID=%d, UpdatedAt=%v", cursor.NmID, cursor.UpdatedAt)
						nomenclatureResponse, err := d.GetNomenclatures(request.Settings{
							Sort:   sort,
							Filter: filter,
							Cursor: cursor,
						}, locale)
						if err != nil {
							log.Printf("Failed to get nomenclatures: %s", err)
							mu.Unlock()
							return
						}

						numItems := len(nomenclatureResponse.Data)
						if numItems == 0 {
							log.Printf("No more data to process. Finishing this job ..")
							mu.Unlock()
							return
						}

						log.Printf("Goroutine %d has done the job! (lastNmID=%d, count=%d)", i, nomenclatureResponse.Data[len(nomenclatureResponse.Data)-1].NmID, numItems)
						cursor.UpdatedAt = nomenclatureResponse.Data[len(nomenclatureResponse.Data)-1].UpdatedAt
						cursor.NmID = nomenclatureResponse.Data[len(nomenclatureResponse.Data)-1].NmID
						log.Printf("Cursor state after request: NmID=%d, UpdatedAt=%v", cursor.NmID, cursor.UpdatedAt)
						totalProcessed += numItems // Увеличиваем общее количество обработанных товаров
						mu.Unlock()

						for _, nomenclature := range nomenclatureResponse.Data {
							dataChan <- nomenclature
						}

						// Проверка: если общее количество обработанных товаров меньше лимита, то завершить обработку
						if totalProcessed >= limit {
							log.Printf("Total processed items (%d) have reached or exceeded the Limit (%d)", totalProcessed, limit)
							return
						}

						if numItems < packetLimit {
							log.Printf("Processed items (%d) are less than the packetLimit (%d), repeating request", numItems, packetLimit)
							continue
						}

						break
					}
				}
			}
		}(i)
	}

	for _, limit := range packetSizes {
		limitChan <- limit
		log.Printf("Sent limit %d to limitChan", limit)
	}
	close(limitChan)
	wg.Wait()

	dataMap := make(map[int]response.Nomenclature)
	data.Range(func(key, value interface{}) bool {
		dataMap[key.(int)] = value.(response.Nomenclature)
		return true
	})

	log.Printf("Total items processed: %d. Successes: %d. Errors: %d", len(dataMap), len(dataMap)-totalErrors, totalErrors)

	return dataMap, nil
}

/*
Возвращает число обновленных(добавленных) карточек
*/
func (d *NomenclatureService) UploadToDb(settings request.Settings, locale string) (int, error) {
	log.Printf("Updating wildberries.nomenclatures")
	log.SetPrefix("NM UPDATER | ")
	updated := 0

	// Получаем существующие номенклатуры из БД
	existIDs, err := d.GetDBNomenclatures()
	if err != nil {
		return updated, err
	}
	client := clients.NewGlobalIDsClient("http://localhost:8081")

	// Инициализируем мапу globalIDsFromDBMap
	globalIDsFromDB, err := client.FetchGlobalIDs()
	if err != nil {
		log.Fatalf("Error fetching Global IDs: %s", err)
	}
	globalIDsFromDBMap := make(map[int]struct{}, len(globalIDsFromDB))
	for _, globalID := range globalIDsFromDB {
		globalIDsFromDBMap[globalID] = struct{}{}
	}

	var limit int
	if settings.Cursor.Limit > len(globalIDsFromDBMap) {
		limit = len(globalIDsFromDBMap)
	} else {
		limit = settings.Cursor.Limit
	}

	nmsFromWb, err := d.GetNomenclaturesWithLimitConcurrency(limit, locale)
	var uploadPacket []interface{}
	for id, nm := range nmsFromWb {
		// есть ли такой товар в базе wholesaler.products
		if _, ok := globalIDsFromDBMap[id]; !ok {
			continue
		}
		if _, ok := existIDs[id]; ok {
			continue
		}
		uploadPacket = append(uploadPacket, id, nm.NmID, nm.ImtID,
			nm.NmUUID, nm.VendorCode, nm.SubjectID,
			nm.Brand, nm.CreatedAt, nm.UpdatedAt)
	}

	log.Printf("Found (%d) new nomenclatures", len(uploadPacket))
	if len(uploadPacket) > 0 {
		err = d.insertBatchNomenclatures(uploadPacket)
		if err != nil {
			return -1, err
		}
		log.Printf("Successfully updates nomenclatures in db")
	} else if len(uploadPacket) == 0 {
		log.Printf("It looks like all the data is up to date")
	}

	log.SetPrefix("")
	return updated, nil
}

func (d *NomenclatureService) insertBatchNomenclatures(batch []interface{}) error {
	query := `
		INSERT INTO wildberries.nomenclatures (global_id, nm_id, imt_id, nm_uuid, vendor_code, subject_id, wb_brand, created_at, updated_at)
		VALUES `

	// Строим запрос со значениями
	valueStrings := []string{}
	for i := 0; i < len(batch)/9; i++ {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
	}

	query += strings.Join(valueStrings, ", ")
	/*
		нужна проверка на то, что добавляемый global_id точно есть в таблице wholesaler.products иначе вылетит паника
		2024-10-28 08:40:27.585 UTC [68] DETAIL:  Key (global_id)=(25268) is not present in table "products".
	*/
	query += `
		ON CONFLICT (global_id) DO UPDATE
		SET nm_id = EXCLUDED.nm_id,
			imt_id = EXCLUDED.imt_id,
			nm_uuid = EXCLUDED.nm_uuid,
			vendor_code = EXCLUDED.vendor_code,
			subject_id = EXCLUDED.subject_id,
			wb_brand = EXCLUDED.wb_brand,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at;
	`

	// Выполняем запрос с батчем параметров
	_, err := d.db.Exec(query, batch...)
	return err
}

func (d *NomenclatureService) GetDBNomenclatures() (map[int]struct{}, error) {
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

func (d *NomenclatureService) initializeCard(nomenclature response.Nomenclature) error {
	dataJson, err := json.Marshal(nomenclature)
	if err != nil {
		return err
	}
	globalID, err := nomenclature.GlobalID()
	if err != nil {
		return err
	}

	cardData := map[string]interface{}{
		"globalID":     globalID,
		"nmID":         nomenclature.NmID,
		"vendorCode":   nomenclature.VendorCode,
		"version_data": dataJson,
	}

	return d.updateService.InitializeCard(cardData)
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
