package responses

import (
	"gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type CategoryResponse struct {
	Data []response.Category `json:"data"`
}
