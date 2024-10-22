package responses

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
)

type ColorResponse struct {
	Data []get.Color `json:"data"`
}
