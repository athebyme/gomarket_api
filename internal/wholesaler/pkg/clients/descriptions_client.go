package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type DescriptionsClient struct {
	ApiURL string
}

func (c DescriptionsClient) FetchDescriptions() (map[int]string, error) {
	log.Printf("Got signal for FetchDescriptions()")
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

	var appellations map[int]string
	if err := json.Unmarshal(body, &appellations); err != nil {
		return nil, err
	}

	return appellations, nil
}

func NewDescriptionsClient(apiURL string) *DescriptionsClient {
	return &DescriptionsClient{ApiURL: apiURL}
}
