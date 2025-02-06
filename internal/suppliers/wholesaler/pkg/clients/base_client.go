package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gomarketplace_api/pkg/logger"
	"io"
	"net/http"
	"time"
)

type BaseClient struct {
	ApiURL string
	log    logger.Logger
	client *http.Client
}

func NewBaseClient(apiURL string, writer io.Writer, logPrefix string) *BaseClient {
	return &BaseClient{
		ApiURL: apiURL,
		log:    logger.NewLogger(writer, logPrefix),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *BaseClient) doRequest(ctx context.Context, method, endpoint string, requestBody interface{}, response interface{}) error {
	c.log.Log(fmt.Sprintf("Got signal for %s", endpoint))

	var bodyBytes []byte
	if requestBody != nil {
		var err error
		bodyBytes, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.ApiURL, endpoint), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return fmt.Errorf("request was cancelled: %w", ctx.Err())
		default:
			return fmt.Errorf("failed to execute request: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
