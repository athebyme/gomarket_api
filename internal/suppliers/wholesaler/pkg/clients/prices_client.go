package clients

import (
	"context"
	"io"
	"net/http"
)

type PriceClient struct {
	BaseClient
}

func NewPriceClient(apiURL string, writer io.Writer) *PriceClient {
	return &PriceClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[WS PriceClient]"),
	}
}

func (c *PriceClient) Fetch(ctx context.Context, requestBody interface{}) (interface{}, error) {
	var prices map[int]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/api/price", requestBody, &prices)
	return prices, err
}
