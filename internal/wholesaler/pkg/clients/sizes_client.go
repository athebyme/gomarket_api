package clients

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
)

type SizesClient struct {
	ApiURL string
	logger logger.Logger
}

func (c *SizesClient) FetchSizes() (map[int][]models.SizeWrapper, error) {
	c.logger.Log("Got signal for FetchSizes()")
	resp, err := http.Get(fmt.Sprintf("%s/api/sizes", c.ApiURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Sizes, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sizes map[int][]models.SizeWrapper
	if err := json.Unmarshal(body, &sizes); err != nil {
		return nil, err
	}

	return sizes, nil
}

func NewSizesClient(apiURL string, writer io.Writer) *SizesClient {
	_log := logger.NewLogger(writer, "[SizesClient]")
	return &SizesClient{ApiURL: apiURL, logger: _log}
}
