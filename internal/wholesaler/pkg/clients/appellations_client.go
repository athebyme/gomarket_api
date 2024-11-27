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

type AppellationsClient struct {
	ApiURL string
	log    logger.Logger
}

func NewAppellationsClient(apiURL string, writer io.Writer) *AppellationsClient {
	_log := logger.NewLogger(writer, "[WS AppellationClient]")

	return &AppellationsClient{ApiURL: apiURL, log: _log}
}

func (c AppellationsClient) FetchAppellations(requestBody interface{}) (map[int]interface{}, error) {
	c.log.Log("Got signal for FetchAppellations()")

	// Преобразуем requestBody в JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Создаем HTTP-запрос с методом POST
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/appellations", c.ApiURL), bytes.NewBuffer(bodyBytes))
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
		return nil, fmt.Errorf("failed to fetch Appellations, status code: %d", resp.StatusCode)
	}

	// Читаем тело ответа
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Десериализуем JSON-ответ в map[int]interface{}
	var appellations map[int]interface{}
	if err := json.Unmarshal(respBody, &appellations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return appellations, nil
}
