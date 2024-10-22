package responses

import "gomarketplace_api/internal/wildberries/internal/business/models"

type CharacteristicsResponse struct {
	Data []models.Characteristic `json:"data"`
}

func (cr *CharacteristicsResponse) FilterPopularity() *CharacteristicsResponse {

	popularCount := 0
	for _, c := range cr.Data {
		if c.Popular {
			popularCount++
		}
	}

	filtered := CharacteristicsResponse{}
	filtered.Data = make([]models.Characteristic, popularCount)
	ind := 0
	for _, c := range cr.Data {
		if c.Popular {
			filtered.Data[ind] = c
			ind++
		}
	}
	return &filtered
}
