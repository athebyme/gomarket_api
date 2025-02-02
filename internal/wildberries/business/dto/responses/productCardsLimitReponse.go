package responses

import (
	"gomarketplace_api/internal/wildberries/business/models/get"
)

type ProductCardsLimitResponse struct {
	Data get.ProductCardsLimit `json:"data"`
}
