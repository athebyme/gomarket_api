package builder

import "gomarketplace_api/internal/wildberries/internal/business/models/dto/response"

type ProductCardBuilder struct {
	card *response.Nomenclature
}

func NewProductCardBuilder() *ProductCardBuilder {
	return &ProductCardBuilder{card: &response.Nomenclature{}}
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
	b.card. = desc
	return b
}

func (b *ProductCardBuilder) WithDimensions(length, width, height int) *ProductCardBuilder {
	b.card.Dimensions = &models.Dimensions{
		Length: length,
		Width:  width,
		Height: height,
	}
	return b
}

func (b *ProductCardBuilder) WithCharacteristics(characteristics []responses.Characteristic) *ProductCardBuilder {
	for _, ch := range characteristics {
		b.card.Characteristics = append(b.card.Characteristics, models.Characteristic{Name: ch.Name, Value: ch.Value})
	}
	return b
}

func (b *ProductCardBuilder) WithSizes(sizes []responses.Size) *ProductCardBuilder {
	for _, size := range sizes {
		b.card.Sizes = append(b.card.Sizes, models.Size{
			WbSize:  size.WbSize,
			TechSize: size.TechSize,
			Barcode:  size.Barcode,
		})
	}
	return b
}

func (b *ProductCardBuilder) Build() *models.ProductCard {
	return b.card
}