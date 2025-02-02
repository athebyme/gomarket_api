package get

import (
	"encoding/json"
	response2 "gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type WildberriesCard struct {
	NmID            int                        `json:"nmID"`
	VendorCode      string                     `json:"vendorCode"`
	Brand           string                     `json:"brand"`
	Title           string                     `json:"title"`
	Description     string                     `json:"description"`
	Dimensions      response2.DimensionWrapper `json:"dimensions"`
	Characteristics []response2.Charc          `json:"characteristics"`
	Sizes           []response2.Size           `json:"sizes"`
}

func (c *WildberriesCard) FromNomenclature(n response2.Nomenclature) *WildberriesCard {
	card := WildberriesCard{}

	card.NmID = n.NmID
	card.VendorCode = n.VendorCode
	card.Brand = n.Brand
	card.Title = n.Title
	card.Characteristics = n.Characteristics
	card.Sizes = n.Sizes
	card.Dimensions = *n.Dimensions.Unwrap()

	return &card
}

func (c *WildberriesCard) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}
