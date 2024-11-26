package clients

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
)

type DescriptionsClient struct {
	ApiURL string
	log    logger.Logger
}

func NewDescriptionsClient(apiURL string, writer io.Writer) *DescriptionsClient {
	_log := logger.NewLogger(writer, "[WS DescriptionsClient]")

	return &DescriptionsClient{ApiURL: apiURL, log: _log}
}

func (c DescriptionsClient) FetchDescriptions() (map[int]interface{}, error) {
	c.log.Log("Got signal for FetchDescriptions()")
	resp, err := http.Get(fmt.Sprintf("%s/api/descriptions", c.ApiURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Descriptions, status code: %d", resp.StatusCode)
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
