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

type DescriptionsClient struct {
	ApiURL string
	log    logger.Logger
}

func NewDescriptionsClient(apiURL string, writer io.Writer) *DescriptionsClient {
	_log := logger.NewLogger(writer, "[WS DescriptionsClient]")

	return &DescriptionsClient{ApiURL: apiURL, log: _log}
}

func (c DescriptionsClient) FetchDescriptions(requestBody interface{}) (map[int]interface{}, error) {
	c.log.Log("Got signal for FetchDescriptions()")

	// Сериализуем requestBody в JSON
	bodyData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Создаем HTTP-запрос с телом
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/descriptions", c.ApiURL), bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус-код
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch descriptions, status code: %d", resp.StatusCode)
	}

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Декодируем JSON-ответ в map[int]interface{}
	var descriptions map[int]interface{}
	if err := json.Unmarshal(body, &descriptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return descriptions, nil
}
