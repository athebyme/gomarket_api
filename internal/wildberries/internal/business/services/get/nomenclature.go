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
	"time"
)

type NomenclatureUpdateGetter struct {
	db            *sql.DB
	updateService get.UpdateService
}

func NewNomenclatureUpdateGetter(db *sql.DB) *NomenclatureUpdateGetter {
	return &NomenclatureUpdateGetter{db: db}
}

const postNomenclature = "https://content-api.wildberries.ru/content/v2/get/cards/list"

func (d *NomenclatureUpdateGetter) GetNomenclature(settings request.Settings, locale string) (*responses.NomenclatureResponse, error) {
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

func (d *NomenclatureUpdateGetter) GetNomenclatureWithLimit(cardCount int, locale string) (*responses.NomenclatureResponse, error) {
	if cardCount <= 0 {
		return nil, fmt.Errorf("cardCount must be greater than 0: %d", cardCount)
	}
	if cardCount > 100 {
		return nil, fmt.Errorf("cardCount must be less than 100 in one request: %d", cardCount)
	}

	url := postNomenclature
	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: 20 * time.Second}

	settings := request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: cardCount},
	}

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

func NewDbNomenclatureUpdater(updateService get.UpdateService, db *sql.DB) *NomenclatureUpdateGetter {
	return &NomenclatureUpdateGetter{updateService: updateService, db: db}
}

/*
Возвращает число обновленных(добавленных) карточек
*/
func (d *NomenclatureUpdateGetter) UpdateNomenclature(settings request.Settings, locale string) (int, error) {
	log.Printf("Updating wildberries.nomenclatures")
	const batchSize = 5
	updated := 0
	var batch []interface{}

	// Получаем существующие номенклатуры из БД
	existIDs, err := d.getAllNomenclatures()
	if err != nil {
		return updated, err
	}

	var r, b int
	if settings.Cursor.Limit < 20 {
		r, b = 5, 5
	} else if settings.Cursor.Limit < 50 {
		r, b = 2, 2
	} else {
		r, b = 1, 1
	}
	limiter := rate.NewLimiter(rate.Limit(r), b)
	client := clients.NewGlobalIDsClient("http://localhost:8081")

	// список всех global ids в wholesaler.products
	globalIDs, err := client.FetchGlobalIDs()
	if err != nil {
		log.Fatalf("Error fetching Global IDs: %s", err)
	}

	globalIDsMap := make(map[int]struct{}, len(globalIDs))

	for _, globalID := range globalIDs {
		globalIDsMap[globalID] = struct{}{}
	}

	for settings.Cursor.Pagination.TotalCards() >= settings.Cursor.Limit {
		if err := limiter.Wait(context.Background()); err != nil {
			return updated, err
		}

		// Получаем данные номенклатур из API
		nomenclatureResponse, err := d.GetNomenclature(settings, locale)
		if err != nil {
			return updated, fmt.Errorf("failed to get nomenclatures: %w", err)
		}

		// Обрабатываем номенклатуры
		for _, nomenclature := range nomenclatureResponse.Data {
			globalId, err := nomenclature.GlobalID()
			if err != nil {
				return updated, fmt.Errorf("failed to get global_id: %w", err)
			}
			if _, exists := existIDs[globalId]; exists {
				continue // Пропускаем существующие записи в таблице wildberries.nomenclatures
			}

			if _, exists := globalIDsMap[globalId]; !exists {
				continue // Пропускаем НЕ существующие записи в таблице wholesaler.products по global id
			}
			if globalId == 0 {
				continue
			}

			if err := d.initializeCard(nomenclature); err != nil {
				return updated, fmt.Errorf("failed to initialize card in history: %w", err)
			}

			log.Printf("updated=%d", updated)
			batch = append(batch, globalId, nomenclature.NmID, nomenclature.ImtID,
				nomenclature.NmUUID, nomenclature.VendorCode, nomenclature.SubjectID,
				nomenclature.Brand, nomenclature.CreatedAt, nomenclature.UpdatedAt)

			if len(batch) >= batchSize*12 {
				if err := d.insertBatchNomenclatures(batch); err != nil {
					return updated, fmt.Errorf("failed to insert batch: %w", err)
				}
				updated += len(batch) / 12
				batch = batch[:0]
			}
		}

		// Проверяем условие завершения пагинации
		if len(nomenclatureResponse.Data) < settings.Cursor.Limit {
			break
		}

		settings.Cursor.Pagination = nomenclatureResponse.Paginator.GetPaginatorCursor()
	}

	// Вставка оставшегося батча
	if len(batch) > 0 {
		if err := d.insertBatchNomenclatures(batch); err != nil {
			return updated, fmt.Errorf("failed to insert final batch: %w", err)
		}
		updated += len(batch) / 12
	}

	return updated, nil
}

func (d *NomenclatureUpdateGetter) insertBatchNomenclatures(batch []interface{}) error {
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

func (d *NomenclatureUpdateGetter) getAllNomenclatures() (map[int]struct{}, error) {
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

func (d *NomenclatureUpdateGetter) initializeCard(nomenclature response.Nomenclature) error {
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
