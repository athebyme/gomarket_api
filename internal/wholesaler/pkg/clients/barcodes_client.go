package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gomarketplace_api/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
)

type BarcodesClient struct {
	ApiURL string
	logger logger.Logger
}

func NewBarcodesClient(apiURL string, writer io.Writer) *BarcodesClient {
	_log := logger.NewLogger(writer, "[WS BrandClient]")
	return &BarcodesClient{
		ApiURL: apiURL,
		logger: _log,
	}
}

func (c BarcodesClient) FetchBarcodes(requestBody interface{}) (map[int]interface{}, error) {
	c.logger.Log("Got signal for FetchBarcodes()")

	// Преобразуем requestBody в JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Создаем HTTP-запрос с методом POST
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/barcodes", c.ApiURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки (например, Content-Type)
	req.Header.Set("Content-Type", "application/json")

	// Отправляем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Brands, status code: %d", resp.StatusCode)
	}

	// Читаем тело ответа
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Десериализуем JSON-ответ в map[int]interface{}
	var barcodes map[int]interface{}
	if err := json.Unmarshal(respBody, &barcodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return barcodes, nil
}
