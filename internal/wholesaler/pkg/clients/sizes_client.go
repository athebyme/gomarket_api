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

type SizesClient struct {
	ApiURL string
	logger logger.Logger
}

func (c *SizesClient) FetchSizes(requestBody interface{}) (map[int][]interface{}, error) {
	c.logger.Log("Got signal for FetchSizes()")

	// Преобразуем requestBody в JSON
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Создаём HTTP-запрос с телом
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/sizes", c.ApiURL), bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Отправляем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Проверяем статус код ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Sizes, status code: %d", resp.StatusCode)
	}

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Парсим JSON-ответ
	var sizes map[int][]interface{}
	if err := json.Unmarshal(body, &sizes); err != nil {
		return nil, err
	}

	return sizes, nil
}

func NewSizesClient(apiURL string, writer io.Writer) *SizesClient {
	_log := logger.NewLogger(writer, "[SizesClient]")
	return &SizesClient{ApiURL: apiURL, logger: _log}
}
