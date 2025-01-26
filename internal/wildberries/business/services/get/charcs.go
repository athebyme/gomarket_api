package get

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"gomarketplace_api/internal/wildberries/business/dto/responses"
	"gomarketplace_api/internal/wildberries/business/services"
	"log"
	"net/http"
	"strings"
	"time"
)

const characteristicsURL = "https://content-api.wildberries.ru/content/v2/object/charcs/%d"

type CharacteristicsEngine struct {
	services.AuthEngine
}

func NewCharacteristicService(auth services.AuthEngine) *CharacteristicsEngine {
	return &CharacteristicsEngine{
		auth,
	}
}

func (s *CharacteristicsEngine) GetItemCharcs(subjectID int, locale string) (*responses.CharacteristicsResponse, error) {
	url := fmt.Sprintf(characteristicsURL, subjectID)
	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	s.SetApiKey(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var characteristicsResp responses.CharacteristicsResponse
	if err := json.NewDecoder(resp.Body).Decode(&characteristicsResp); err != nil {
		return nil, err
	}

	return &characteristicsResp, nil
}

type UpdateDBCharcs struct {
	db *sql.DB
	CharacteristicsEngine
}

func NewUpdateDBCharcs(db *sql.DB, service CharacteristicsEngine) *UpdateDBCharcs {
	return &UpdateDBCharcs{db: db, CharacteristicsEngine: service}
}

func (d *UpdateDBCharcs) UpdateDBCharcs(subjectIDs []int) (int, error) {
	const batchSize = 5 // Размер пачки для вставки
	updated := 0
	var batch []interface{}
	existCharc, err := d.getAllCharcsInDb()
	if err != nil {
		return updated, err
	}
	limiter := rate.NewLimiter(5, 5)

	log.Printf("looking for (subjectID=%v)", subjectIDs)
	for _, subjectID := range subjectIDs {

		log.Printf("%d", subjectID)
		if err := limiter.Wait(context.Background()); err != nil {
			return -1, err
		}
		response, err := d.GetItemCharcs(subjectID, "")
		if err != nil {
			return -1, err
		}
		data := response.Data
		for _, charc := range data {
			// Добавляем данные в батч
			if _, ok := existCharc[charc.ID]; ok {
				log.Printf("(subjectID=%d) (charcID=%d) already exists", subjectID, charc.ID)
				continue
			}
			batch = append(batch, charc.ID, charc.Name, charc.Required, subjectID, charc.UnitName, charc.MaxCount, charc.Popular, charc.CharcType)

			// Если собрали полную пачку, то отправляем её в базу
			if len(batch)/8 >= batchSize {
				if err := d.insertBatch(batch); err != nil {
					return updated, err
				}
				updated += batchSize
				batch = batch[:0] // Очищаем батч после отправки
			}
			existCharc[charc.ID] = struct{}{}
		}

	}

	// Если остались записи в батче, отправляем их
	if len(batch) > 0 {
		if err := d.insertBatch(batch); err != nil {
			return updated, err
		}
		updated += len(batch) / 8
	}

	return updated, nil
}

func (d *UpdateDBCharcs) insertBatch(batch []interface{}) error {
	query := `
		INSERT INTO wildberries.characteristics (charc_id, name, required, subject_id, unit_name, max_count, popular, charc_type)
		VALUES `

	// Строим запрос со значениями
	valueStrings := []string{}
	for i := 0; i < len(batch)/8; i++ {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8))
	}

	query += strings.Join(valueStrings, ", ")
	query += `
		ON CONFLICT (charc_id) DO UPDATE
		SET name = EXCLUDED.name,
			required = EXCLUDED.required,
			subject_id = EXCLUDED.subject_id,
			unit_name = EXCLUDED.unit_name,
			max_count = EXCLUDED.max_count,
			popular = EXCLUDED.popular,
			charc_type = EXCLUDED.charc_type;
	`

	// Выполняем запрос с батчем параметров
	_, err := d.db.Exec(query, batch...)
	return err
}

func (d *UpdateDBCharcs) getAllCharcsInDb() (map[int]struct{}, error) {
	// запрос для получения списка category_id
	query := `SELECT charc_id FROM wildberries.characteristics`

	// создаем срез для хранения category_id
	charcIDs := make(map[int]struct{}, 1)
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
		charcIDs[catID] = struct{}{}
	}

	// проверяем ошибки после цикла rows.Next()
	if err := rows.Err(); err != nil {
		return map[int]struct{}{}, fmt.Errorf("ошибка чтения строк: %w", err)
	}

	return charcIDs, nil
}

/*
Возвращает количество обновленных/записанных строк
*/
func UploadItemCharcs(db *sql.DB, updatedCharacteristics map[int]string) (int, error) {
	// запрос для получения списка category_id
	query := `SELECT category_id FROM wildberries.categories`

	// создаем срез для хранения category_id
	var catIDs []int
	rows, err := db.Query(query)
	if err != nil {
		return -1, fmt.Errorf("ошибка выполнения запроса для категорий: %w", err)
	}
	defer rows.Close()

	// заполняем срез category_id из результата запроса
	for rows.Next() {
		var catID int
		if err := rows.Scan(&catID); err != nil {
			return -1, fmt.Errorf("ошибка сканирования cat_id: %w", err)
		}
		catIDs = append(catIDs, catID)
	}

	// проверяем ошибки после цикла rows.Next()
	if err := rows.Err(); err != nil {
		return -1, fmt.Errorf("ошибка чтения строк: %w", err)
	}

	// подготавливаем запросы для обновления и вставки характеристик
	updateQuery := `UPDATE wildberries.characteristics SET characteristic_value = $1 WHERE category_id = $2`
	insertQuery := `INSERT INTO wildberries.characteristics (category_id, characteristic_value) VALUES ($1, $2)`

	var updatedCount, insertedCount int
	for _, catID := range catIDs {
		// получаем новое значение для обновления, если оно есть в карте
		if newValue, exists := updatedCharacteristics[catID]; exists {
			// выполняем попытку обновления
			result, err := db.Exec(updateQuery, newValue, catID)
			if err != nil {
				return -1, fmt.Errorf("ошибка обновления характеристики для category_id %d: %w", catID, err)
			}

			// проверяем количество затронутых строк
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return -1, fmt.Errorf("ошибка проверки количества затронутых строк: %w", err)
			}

			if rowsAffected == 0 {
				// если строка не обновлена, выполняем вставку
				_, err := db.Exec(insertQuery, catID, newValue)
				if err != nil {
					return -1, fmt.Errorf("ошибка вставки характеристики для category_id %d: %w", catID, err)
				}
				insertedCount++
			} else {
				updatedCount++
			}
		}
	}

	return len(updatedCharacteristics), nil
}
