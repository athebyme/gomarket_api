package responses

import "gomarketplace_api/internal/wildberries/internal/business/models"

type ProductCardsLimitResponse struct {
	Data models.ProductCardsLimit `json:"data"`
}
