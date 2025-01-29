package clients

import (
	"context"
	"io"
	"net/http"
)

type ImageClient struct {
	BaseClient
}

type ImageRequest struct {
	ProductIDs []int `json:"productIDs"`
	Censored   bool  `json:"censored"`
}

func NewImageClient(apiURL string, writer io.Writer) *ImageClient {
	return &ImageClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[WS ImageClient]"),
	}
}

func (c *ImageClient) Fetch(ctx context.Context, requestBody interface{}) (interface{}, error) {
	var images map[int][]string
	err := c.doRequest(ctx, http.MethodPost, "/api/media", requestBody, &images)
	return images, err
}
