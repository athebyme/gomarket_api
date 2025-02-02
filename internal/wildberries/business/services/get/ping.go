package get

import (
	"encoding/json"
	"errors"
	"gomarketplace_api/internal/wildberries/business/dto/responses"
	"gomarketplace_api/internal/wildberries/business/services"
	"net/http"
	"time"
)

const wildberriesPingURL = "https://common-api.wildberries.ru/ping"

type PingEngine struct {
	services.AuthEngine
}

func NewPingEngine(auth services.AuthEngine) *PingEngine {
	return &PingEngine{auth}
}

func (pe *PingEngine) Ping() (*responses.Ping, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", wildberriesPingURL, nil)
	if err != nil {
		return nil, err
	}

	pe.SetApiKey(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	var pingResp responses.Ping
	if err := json.NewDecoder(resp.Body).Decode(&pingResp); err != nil {
		return nil, err
	}

	return &pingResp, nil
}
