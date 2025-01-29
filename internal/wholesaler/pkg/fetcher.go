package pkg

import (
	"context"
	"fmt"
	"gomarketplace_api/pkg/logger"
	"sync"
	"time"
)

// FetcherId используется для создания фетчера в цепочке
type FetcherId struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type FetcherChain struct {
	fetchers      map[string]Fetcher
	fetcherIds    map[int]FetcherId
	log           logger.Logger
	existingNames map[string]bool
	nextID        int
	mu            sync.Mutex
}

func NewFetcherChain(log logger.Logger) *FetcherChain {
	return &FetcherChain{
		fetchers:      make(map[string]Fetcher),
		fetcherIds:    make(map[int]FetcherId),
		log:           log,
		existingNames: make(map[string]bool),
		nextID:        1,
	}
}

func (fc *FetcherChain) Register(fetcherName string, fetcher Fetcher) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if fetcher == nil {
		err := fmt.Errorf("fetcher is nil for name '%s'", fetcherName)
		fc.log.Log(fmt.Sprintf("Error registering fetcher: %v", err))
		return err
	}

	if err := validateFetcherName(fetcherName, fc.existingNames); err != nil {
		fc.log.Log(fmt.Sprintf("Error validating fetcher name: %v", err))
		return err
	}

	newID := fc.nextID
	fc.nextID++

	fetcherId := FetcherId{
		Id:   newID,
		Name: fetcherName,
	}

	fc.fetchers[fetcherName] = fetcher
	fc.fetcherIds[newID] = fetcherId
	fc.existingNames[fetcherName] = true

	fc.log.Log(fmt.Sprintf("Successfully registered fetcher: id=%d, name=%s", newID, fetcherName))
	return nil
}

func (fc *FetcherChain) GetFetcherByName(name string) (Fetcher, error) {
	fetcher, ok := fc.fetchers[name]
	if !ok {
		err := fmt.Errorf("fetcher with name '%s' not found", name)
		fc.log.Log(fmt.Sprintf("Error getting fetcher: %v", err))
		return nil, err
	}
	return fetcher, nil
}

func (fc *FetcherChain) GetFetcherById(id int) (Fetcher, error) {
	fetcherId, ok := fc.fetcherIds[id]
	if !ok {
		err := fmt.Errorf("fetcher with id %d not found", id)
		fc.log.Log(fmt.Sprintf("Error getting fetcher: %v", err))
		return nil, err
	}
	return fc.GetFetcherByName(fetcherId.Name)
}

func (fc *FetcherChain) Fetch(ctx context.Context, fetcherName string, request interface{}) (interface{}, error) {
	fetcher, ok := fc.fetchers[fetcherName]
	if !ok {
		err := fmt.Errorf("fetcher '%s' not found", fetcherName)
		fc.log.Log(fmt.Sprintf("Error fetching data: %v", err))
		return nil, err
	}

	fc.log.Log(fmt.Sprintf("Fetching data using fetcher: %s", fetcherName))
	return fetcher.Fetch(ctx, request)
}

func (fc *FetcherChain) FetchWithTimeout(ctx context.Context, fetcherName string, request interface{}, timeout time.Duration) (interface{}, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	fetcher, ok := fc.fetchers[fetcherName]
	if !ok {
		err := fmt.Errorf("fetcher '%s' not found", fetcherName)
		fc.log.Log(fmt.Sprintf("Error fetching data: %v", err))
		return nil, err
	}

	fc.log.Log(fmt.Sprintf("Fetching data using fetcher: %s with timeout %v", fetcherName, timeout))
	return fetcher.Fetch(ctxWithTimeout, request)
}

func validateFetcherName(fetcherName string, existingNames map[string]bool) error {
	if fetcherName == "" {
		return fmt.Errorf("fetcher name cannot be empty")
	}
	if existingNames[fetcherName] {
		return fmt.Errorf("fetcher with name '%s' already exists", fetcherName)
	}
	return nil
}
