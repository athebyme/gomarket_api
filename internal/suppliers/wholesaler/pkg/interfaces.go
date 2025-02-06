package pkg

import "context"

type Fetcher interface {
	Fetch(ctx context.Context, request interface{}) (interface{}, error)
}
