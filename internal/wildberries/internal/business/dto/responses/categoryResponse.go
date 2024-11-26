package responses

import "gomarketplace_api/internal/wildberries/internal/business/models/dto/response"

type CategoryResponse struct {
	Data []response.Category `json:"data"`
}
