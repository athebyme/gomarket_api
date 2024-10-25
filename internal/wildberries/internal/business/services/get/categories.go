package get

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"net/http"
	"net/url"
	"time"
)

const categoriesUrl = "https://content-api.wildberries.ru/content/v2/object/all"

// GetCategories запрашивает категории с указанными параметрами: имя, локаль, лимит, смещение и идентификатор родителя.
func GetCategories(name, locale string, limit, offset, parentID int) (*responses.CategoryResponse, error) {
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
	services.SetAuthorizationHeader(req)

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
