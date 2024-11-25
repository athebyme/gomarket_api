package clients

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
)

type BrandsClient struct {
	ApiURL string
	logger logger.Logger
}

func NewBrandsClient(apiURL string, writer io.Writer) *BrandsClient {
	_log := logger.NewLogger(writer, "[WS BrandClient]")
	return &BrandsClient{
		ApiURL: apiURL,
		logger: _log,
	}
}

func (c BrandsClient) FetchBrands() (map[int]interface{}, error) {
	c.logger.Log("Got signal for FetchBrands()")
	resp, err := http.Get(fmt.Sprintf("%s/api/brands", c.ApiURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Brands, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var appellations map[int]interface{}
	if err := json.Unmarshal(body, &appellations); err != nil {
		return nil, err
	}

	return appellations, nil
}
