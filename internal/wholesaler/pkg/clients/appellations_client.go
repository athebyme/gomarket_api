package clients

import (
	"context"
	"io"
	"net/http"
)

type AppellationsFetcher struct {
	BaseClient
}

func NewAppellationsFetcher(apiURL string, writer io.Writer) *AppellationsFetcher {
	return &AppellationsFetcher{
		BaseClient: *NewBaseClient(apiURL, writer, "[WS BarcodesClient]"),
	}
}

func (c AppellationsFetcher) Fetch(ctx context.Context, request interface{}) (interface{}, error) {
	var appellations map[int]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/api/appellations", request, &appellations)
	if err != nil {
		return nil, err
	}

	return appellations, nil
}
