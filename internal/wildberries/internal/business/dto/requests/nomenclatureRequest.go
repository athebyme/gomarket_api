package requests

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
)

type NomenclatureRequest struct {
	Settings request.Settings `json:"settings"`
}
