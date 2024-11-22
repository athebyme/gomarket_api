package clients

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/business"
	"io/ioutil"
	"log"
	"net/http"
)

type PriceClient struct {
	ApiURL string
}

func (c PriceClient) FetchPrices() (map[int]business.PriceResult, error) {
	log.Printf("Got signal for FetchPrices()")
	resp, err := http.Get(fmt.Sprintf("%s/api/prices", c.ApiURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Prices, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var prices map[int]business.PriceResult
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, err
	}

	return prices, nil
}

func NewPriceClient(apiURL string) *PriceClient {
	return &PriceClient{ApiURL: apiURL}
}
