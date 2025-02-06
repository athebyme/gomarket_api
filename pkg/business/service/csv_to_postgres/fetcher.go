package csv_to_postgres

import (
	"fmt"
	"io"
	"net/http"
)

// Fetcher определяет интерфейс для получения данных по URL
type Fetcher interface {
	Fetch(url string) (io.ReadCloser, error)
}

type HTTPFetcher struct {
	Client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		Client: &http.Client{},
	}
}

func (f *HTTPFetcher) Fetch(url string) (io.ReadCloser, error) {
	resp, err := f.Client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
