package get

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"net/http"
	"net/url"
	"time"
)

type CategoriesEngine struct {
	services.AuthEngine
}

func NewCategoriesService(auth services.AuthEngine) *CategoriesEngine {
	return &CategoriesEngine{
		auth,
	}
}

const categoriesUrl = "https://content-api.wildberries.ru/content/v2/object/all"

// GetCategories запрашивает категории с указанными параметрами: имя, локаль, лимит, смещение и идентификатор родителя.
func (s *CategoriesEngine) GetCategories(name, locale string, limit, offset, parentID int) (*responses.CategoryResponse, error) {
	// Формируем URL с параметрами
	url, err := buildCategoriesURL(locale, limit, offset, parentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка формирования URL: %w", err)
	}

	// Создаем HTTP клиент с тайм-аутом
	client := &http.Client{Timeout: 10 * time.Second}

	// Создаем новый запрос
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	// Устанавливаем заголовок авторизации
	s.SetApiKey(req)

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неожиданный статус ответа: %s", resp.Status)
	}

	// Декодируем тело ответа в структуру CategoryResponse
	var categoriesResponse responses.CategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&categoriesResponse); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	return &categoriesResponse, nil
}

type DBCategories struct {
	db *sql.DB
}

func NewDBCategories(db *sql.DB) *DBCategories {
	return &DBCategories{db}
}

func (c *DBCategories) Categories() ([]response.Category, error) {
	query := `
		SELECT category_id, parent_category_id, category, parent_category_name
		FROM wildberries.categories
	`

	// Выполняем запрос к базе данных
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса категорий: %w", err)
	}
	defer rows.Close()

	// Инициализируем срез для хранения категорий
	var categories []response.Category

	// Проходим по результатам запроса и сканируем данные в структуру Category
	for rows.Next() {
		var category response.Category
		if err := rows.Scan(&category.SubjectID, &category.ParentID, &category.SubjectName, &category.ParentName); err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных категории: %w", err)
		}
		categories = append(categories, category)
	}

	// Проверяем ошибки после цикла rows.Next()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения строк: %w", err)
	}

	return categories, nil
}

// buildCategoriesURL формирует URL для запроса категорий, добавляя параметры локали, лимита, смещения и родительского ID.
func buildCategoriesURL(locale string, limit, offset, parentID int) (string, error) {
	params := url.Values{}

	if locale != "" {
		params.Add("locale", locale)
	}
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", offset))
	}
	if parentID > 0 {
		params.Add("parentID", fmt.Sprintf("%d", parentID))
	}

	// Добавляем параметры к базовому URL
	urlWithParams := fmt.Sprintf("%s?%s", categoriesUrl, params.Encode())
	return urlWithParams, nil
}
