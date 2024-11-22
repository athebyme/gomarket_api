package builder

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
)

type CreateCardBuilder struct {
	brand           string
	title           string
	description     string
	vendorCode      string
	dimensions      response.DimensionWrapper
	sizes           []response.Size
	characteristics []response.CharcWrapper
}

func (b *CreateCardBuilder) WithBrand(brand string) *CreateCardBuilder {
	b.brand = brand
	return b
}
func (b *CreateCardBuilder) WithTitle(title string) *CreateCardBuilder {
	b.title = title
	return b
}
func (b *CreateCardBuilder) WithDescription(description string) *CreateCardBuilder {
	b.description = description
	return b
}
func (b *CreateCardBuilder) WithVendorCode(vendorCode string) *CreateCardBuilder {
	b.vendorCode = vendorCode
	return b
}
func (b *CreateCardBuilder) WithDimensions(dimensions response.Dimensions) *CreateCardBuilder {
	b.dimensions = *dimensions.Unwrap()
	return b
}
func (b *CreateCardBuilder) WithDimensionWrapper(dimensions response.DimensionWrapper) *CreateCardBuilder {
	b.dimensions = dimensions
	return b
}
func (b *CreateCardBuilder) WithSizes(sizes []response.Size) *CreateCardBuilder {
	b.sizes = sizes
	return b
}
func (b *CreateCardBuilder) WithCharacteristics(charcs []response.CharcWrapper) *CreateCardBuilder {
	b.characteristics = charcs
	return b
}

func (b *CreateCardBuilder) Build() (interface{}, error) {
	var card *request.CreateCardRequestData
	card = &request.CreateCardRequestData{
		Brand:           b.brand,
		Title:           b.title,
		Description:     b.description,
		VendorCode:      b.vendorCode,
		Dimensions:      b.dimensions,
		Sizes:           b.sizes,
		Characteristics: b.characteristics,
	}
	if err := card.Validate(); err != nil {
		return nil, err
	}
	return card, nil
}
