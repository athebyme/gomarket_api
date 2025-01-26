package responses

import (
	"gomarketplace_api/internal/wildberries/business/models/get"
)

type ColorResponse struct {
	Data []get.Color `json:"data"`
}
