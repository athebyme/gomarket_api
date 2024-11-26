package clients

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
)

type AppellationsClient struct {
	ApiURL string
	log    logger.Logger
}

func NewAppellationsClient(apiURL string, writer io.Writer) *AppellationsClient {
	_log := logger.NewLogger(writer, "[WS AppellationClient]")

	return &AppellationsClient{ApiURL: apiURL, log: _log}
}

func (c AppellationsClient) FetchAppellations() (map[int]interface{}, error) {
	c.log.Log("Got signal for FetchAppellations()")
	resp, err := http.Get(fmt.Sprintf("%s/api/appellations", c.ApiURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Appellations, status code: %d", resp.StatusCode)
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
