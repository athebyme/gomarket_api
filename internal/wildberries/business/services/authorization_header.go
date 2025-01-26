package services

import (
	"net/http"
)

type AuthEngine interface {
	GetApiKey() string
	SetApiKey(request *http.Request)
}

type BearerAuth struct {
	apiKey string
}

func (b *BearerAuth) GetApiKey() string {
	return b.apiKey
}

func (b *BearerAuth) SetApiKey(request *http.Request) {
	request.Header.Set("Authorization", "Bearer "+b.apiKey)
}

func NewBearerAuth(apiKey string) *BearerAuth {
	if apiKey == "" {
		return nil
	}
	return &BearerAuth{apiKey: apiKey}
}
