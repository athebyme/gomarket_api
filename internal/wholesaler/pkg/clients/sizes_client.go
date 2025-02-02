package clients

import (
	"context"
	"io"
	"net/http"
)

type SizesClient struct {
	BaseClient
}

func NewSizesClient(apiURL string, writer io.Writer) *SizesClient {
	return &SizesClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[WS SizesClient]"),
	}
}

func (c *SizesClient) Fetch(ctx context.Context, requestBody interface{}) (interface{}, error) {
	var sizes map[int][]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/api/sizes", requestBody, &sizes)
	return sizes, err
}
