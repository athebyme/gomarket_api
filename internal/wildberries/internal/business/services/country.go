package services

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"net/http"
	"time"
)

const countryUrl = "https://content-api.wildberries.ru/content/v2/directory/countries"

func GetCountries(locale string) (*responses.CountryResponse, error) {
	url := countryUrl
	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	SetAuthorizationHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var countryResponse responses.CountryResponse
	if err := json.NewDecoder(resp.Body).Decode(&countryResponse); err != nil {
		return nil, err
	}

	return &countryResponse, nil
}
