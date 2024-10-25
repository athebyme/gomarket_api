package get

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"net/http"
	"time"
)

const characteristicsURL = "https://content-api.wildberries.ru/content/v2/object/charcs/%d"

func GetItemCharcs(subjectID int, locale string) (*responses.CharacteristicsResponse, error) {
	url := fmt.Sprintf(characteristicsURL, subjectID)

	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	services.SetAuthorizationHeader(req)

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
