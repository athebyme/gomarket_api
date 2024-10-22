package responses

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
)

type CharacteristicsResponse struct {
	Data []get.FullCharcsInfo `json:"data"`
}

func (cr *CharacteristicsResponse) FilterPopularity() *CharacteristicsResponse {

	popularCount := 0
	for _, c := range cr.Data {
		if c.Popular {
			popularCount++
		}
	}

	filtered := CharacteristicsResponse{}
	filtered.Data = make([]get.FullCharcsInfo, popularCount)
	ind := 0
	for _, c := range cr.Data {
		if c.Popular {
			filtered.Data[ind] = c
			ind++
		}
	}
	return &filtered
}
