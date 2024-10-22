package get

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"net/http"
	"time"
)

const productCardsLimitUrl = "https://content-api.wildberries.ru/content/v2/cards/limits"

func GetProductCardsLimit() (*responses.ProductCardsLimitResponse, error) {
	url := productCardsLimitUrl

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

	var prodCardsLim responses.ProductCardsLimitResponse
	if err := json.NewDecoder(resp.Body).Decode(&prodCardsLim); err != nil {
		return nil, err
	}

	return &prodCardsLim, nil
}
