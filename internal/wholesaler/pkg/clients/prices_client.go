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

type PriceClient struct {
	ApiURL string
	log    logger.Logger
}

func (c PriceClient) FetchPrices(requestBody interface{}) (map[int]interface{}, error) {
	c.log.Log("Got signal for FetchPrices()")

	// Сериализация тела запроса в JSON
	bodyData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Отправка POST-запроса с телом
	resp, err := http.Post(fmt.Sprintf("%s/api/price", c.ApiURL), "application/json", bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Prices, status code: %d", resp.StatusCode)
	}

	// Чтение тела ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Десериализация JSON-ответа
	var prices map[int]interface{}
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return prices, nil
}

func NewPriceClient(apiURL string, writer io.Writer) *PriceClient {
	_log := logger.NewLogger(writer, "[WS PriceClient]")
	return &PriceClient{ApiURL: apiURL, log: _log}
}
