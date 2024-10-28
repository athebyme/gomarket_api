package builder

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
)

type ProductCardBuilder struct {
	card         *get.WildberriesCard
	nomenclature *response.Nomenclature
}

func NewProductCardBuilder() *ProductCardBuilder {
	return &ProductCardBuilder{
		card:         &get.WildberriesCard{},
		nomenclature: &response.Nomenclature{}}
}
func (b *ProductCardBuilder) WithNmID(nmID int) *ProductCardBuilder {
	b.card.NmID = nmID
	return b
}

func (b *ProductCardBuilder) WithVendorCode(code string) *ProductCardBuilder {
	b.card.VendorCode = code
	return b
}

func (b *ProductCardBuilder) WithBrand(brand string) *ProductCardBuilder {
	b.card.Brand = brand
	return b
}

func (b *ProductCardBuilder) WithTitle(title string) *ProductCardBuilder {
	b.card.Title = title
	return b
}

func (b *ProductCardBuilder) WithDescription(desc string) *ProductCardBuilder {
	b.card.Description = desc
	return b
}

func (b *ProductCardBuilder) WithNomenclature(n response.Nomenclature) *ProductCardBuilder {
	b.nomenclature = &n
	return b
}
func (b *ProductCardBuilder) CardInfoFromNomenclature() *ProductCardBuilder {
	if b.nomenclature == nil {
		return b
	}
	b.card = b.card.FromNomenclature(*b.nomenclature)
	return b
}
