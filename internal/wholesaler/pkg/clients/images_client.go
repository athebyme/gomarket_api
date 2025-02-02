package clients

import (
	"context"
	"io"
	"net/http"
)

type MediaClient struct {
	BaseClient
}

type ImageRequest struct {
	ProductIDs []int `json:"productIDs"`
	Censored   bool  `json:"censored"`
}

func NewImageClient(apiURL string, writer io.Writer) *MediaClient {
	return &MediaClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[ WS MediaClient ]"),
	}
}

func (c *MediaClient) Fetch(ctx context.Context, requestBody interface{}) (interface{}, error) {
	var images map[int][]string
	err := c.doRequest(ctx, http.MethodPost, "/api/media", requestBody, &images)
	return images, err
}
