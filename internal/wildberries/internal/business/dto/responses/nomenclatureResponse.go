package responses

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
)

type NomenclatureResponse struct {
	Data   []response.Nomenclature `json:"cards"`
	Cursor request.Cursor          `json:"cursor"`
}
