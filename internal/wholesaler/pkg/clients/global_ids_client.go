package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type GlobalIDsClient struct {
	BaseClient
}

func NewGlobalIDsClient(apiURL string, writer io.Writer) *GlobalIDsClient {
	return &GlobalIDsClient{
		BaseClient: *NewBaseClient(apiURL, writer, "[GlobalIDsClient]"),
	}
}

func (f *GlobalIDsClient) Fetch(ctx context.Context, request interface{}) (interface{}, error) {
	var globalIDs []int
	err := f.doRequest(ctx, http.MethodGet, "/api/globalids", request, &globalIDs)
	if err != nil {
		return nil, err
	}
	f.log.Log(fmt.Sprintf("Fetched %d Global IDs", len(globalIDs)))
	return globalIDs, nil
}
