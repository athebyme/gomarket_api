package responses

import "gomarketplace_api/internal/wildberries/internal/business/models"

type ColorResponse struct {
	Data []models.Color `json:"data"`
}
