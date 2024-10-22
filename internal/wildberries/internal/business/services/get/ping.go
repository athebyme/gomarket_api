package get

import (
	"encoding/json"
	"errors"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"net/http"
	"time"
)

const wildberriesPingURL = "https://common-api.wildberries.ru/ping"

func Ping() (*responses.Ping, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", wildberriesPingURL, nil)
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
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	var pingResp responses.Ping
	if err := json.NewDecoder(resp.Body).Decode(&pingResp); err != nil {
		return nil, err
	}

	return &pingResp, nil
}
