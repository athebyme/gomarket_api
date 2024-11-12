package get

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"net/http"
	"time"
)

const colorsUrl = "https://content-api.wildberries.ru/content/v2/directory/colors"

type ColorEngine struct {
	services.AuthEngine
}

func NewColorEngine(auth services.AuthEngine) *ColorEngine {
	return &ColorEngine{auth}
}

func (ce *ColorEngine) GetColors(locale string) (*responses.ColorResponse, error) {
	url := colorsUrl
	if locale != "" {
		url = fmt.Sprintf("%s?locale=%s", url, locale)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	ce.SetApiKey(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var colorsResp responses.ColorResponse
	if err := json.NewDecoder(resp.Body).Decode(&colorsResp); err != nil {
		return nil, err
	}

	return &colorsResp, nil
}
