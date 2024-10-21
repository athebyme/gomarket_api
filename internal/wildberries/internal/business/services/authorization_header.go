package services

import (
	"gomarketplace_api/config"
	"net/http"
)

var apiKey = config.GetMarketplaceConfig()

func SetAuthorizationHeader(request *http.Request) {
	request.Header.Set("Authorization", "Bearer "+apiKey.ApiKey)
}
