package clients

import (
	"context"
	"io"
	"net/http"
)

type BrandsClient struct {
	BaseClient
}

func NewBrandsClient(apiURL string, writer io.Writer) *BrandsClient {
	return &BrandsClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[WS BrandClient]"),
	}
}

func (c *BrandsClient) Fetch(ctx context.Context, requestBody interface{}) (interface{}, error) {
	var brands map[int]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/api/brands", requestBody, &brands)
	return brands, err
}
