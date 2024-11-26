package responses

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
)

type ProductCardsLimitResponse struct {
	Data get.ProductCardsLimit `json:"data"`
}
