package clients

import (
	"context"
	"io"
	"net/http"
)

type BarcodesClient struct {
	BaseClient
}

func NewBarcodesClient(apiURL string, writer io.Writer) *BarcodesClient {
	return &BarcodesClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[WS BarcodesClient]"),
	}
}

func (c *BarcodesClient) Fetch(ctx context.Context, requestBody interface{}) (interface{}, error) {
	var barcodes map[int]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/api/barcodes", requestBody, &barcodes)
	return barcodes, err
}
