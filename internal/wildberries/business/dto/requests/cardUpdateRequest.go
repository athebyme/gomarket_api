package requests

import (
	"gomarketplace_api/internal/wildberries/business/models/get"
)

type CardUpdateRequest struct {
	Data []get.WildberriesCard `json:""`
}
