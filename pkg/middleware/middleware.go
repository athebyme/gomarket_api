package middleware

import "context"

type Middleware func(next func(ctx context.Context, method, endpoint string, requestBody, response interface{}) error) func(ctx context.Context, method, endpoint string, requestBody, response interface{}) error
