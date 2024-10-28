package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GlobalIDsClient struct {
	ApiURL string
}

func NewGlobalIDsClient(apiURL string) *GlobalIDsClient {
	return &GlobalIDsClient{ApiURL: apiURL}
}
func (c *GlobalIDsClient) FetchGlobalIDs() ([]int, error) {
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
