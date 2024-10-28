package get

import "gomarketplace_api/internal/wildberries/internal/business/models/dto/response"

type WildberriesCard struct {
	NmID            int              `json:"nmID"`
	VendorCode      string           `json:"vendorCode"`
	Brand           string           `json:"brand"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Characteristics []response.Charc `json:"characteristics"`
	Sizes           []response.Size  `json:"sizes"`
}

func (c *WildberriesCard) FromNomenclature(n response.Nomenclature) *WildberriesCard {
	v := WildberriesCard{}
	v.NmID = n.NmID
	v.VendorCode = n.VendorCode
	v.Brand = n.Brand
	v.Title = n.Title
	v.Characteristics = n.Characteristics
	v.Sizes = n.Sizes

	return &v
}
