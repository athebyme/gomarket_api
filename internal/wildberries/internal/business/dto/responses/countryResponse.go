package responses

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
)

type CountryResponse struct {
	Data []get.Country `json:"data"`
}
