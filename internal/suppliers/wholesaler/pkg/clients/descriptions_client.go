package clients

import (
	"context"
	"io"
	"net/http"
)

type DescriptionsClient struct {
	BaseClient
}

func NewDescriptionsClient(apiURL string, writer io.Writer) *DescriptionsClient {
	return &DescriptionsClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[WS DescriptionsClient]"),
	}
}

func (c *DescriptionsClient) Fetch(ctx context.Context, requestBody interface{}) (interface{}, error) {
	var descriptions map[int]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/api/descriptions", requestBody, &descriptions)
	return descriptions, err
}
