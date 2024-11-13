package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ImageClient struct {
	ApiURL string
}

func NewImageClient(apiURL string) *ImageClient {
	return &ImageClient{ApiURL: apiURL}
}

type ImageRequest struct {
	ProductIDs []int `json:"productIDs"`
	Censored   bool  `json:"censored"`
}

func (c *GlobalIDsClient) FetchImages(mediaReq ImageRequest) (map[int][]string, error) {
	log.Printf("Got signal for FetchImages()")

	requestBody, err := json.Marshal(mediaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/media", c.ApiURL), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch media, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	mediaMap := make(map[int][]string)
	if err := json.Unmarshal(body, &mediaMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return mediaMap, nil
}
