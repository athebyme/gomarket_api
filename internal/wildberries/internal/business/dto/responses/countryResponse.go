package responses

import "gomarketplace_api/internal/wildberries/internal/business/models"

type CountryResponse struct {
	Data []models.Country `json:"data"`
}
