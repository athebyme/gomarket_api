package clients

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
)

type GlobalIDsClient struct {
	ApiURL string
	log    logger.Logger
}

func NewGlobalIDsClient(apiURL string, writer io.Writer) *GlobalIDsClient {
	_log := logger.NewLogger(writer, "[WS GlobalIDsClient]")

	return &GlobalIDsClient{ApiURL: apiURL, log: _log}
}

func (c *GlobalIDsClient) FetchGlobalIDs() ([]int, error) {
	c.log.Log("Got signal for FetchGlobalIDs()")
	resp, err := http.Get(fmt.Sprintf("%s/api/globalids", c.ApiURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Global IDs, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var globalIDs []int
	if err := json.Unmarshal(body, &globalIDs); err != nil {
		return nil, err
	}

	return globalIDs, nil
}
