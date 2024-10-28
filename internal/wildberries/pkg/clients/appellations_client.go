package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type AppellationsClient struct {
	ApiURL string
}

func (c AppellationsClient) FetchAppellations() (map[int]string, error) {
	log.Printf("Got signal for FetchAppellations()")
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

	var appellations map[int]string
	if err := json.Unmarshal(body, &appellations); err != nil {
		return nil, err
	}

	return appellations, nil
}

func NewAppellationsClient(apiURL string) *AppellationsClient {
	return &AppellationsClient{ApiURL: apiURL}
}
