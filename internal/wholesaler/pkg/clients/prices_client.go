package clients

import (
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

func (c PriceClient) FetchPrices() (map[int]interface{}, error) {
	c.log.Log("Got signal for FetchPrices()")
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

	var prices map[int]interface{}
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, err
	}

	return prices, nil
}

func NewPriceClient(apiURL string, writer io.Writer) *PriceClient {
	_log := logger.NewLogger(writer, "[WS PriceClient]")
	return &PriceClient{ApiURL: apiURL, log: _log}
}
