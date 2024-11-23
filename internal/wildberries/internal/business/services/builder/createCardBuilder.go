package builder

import (
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
)

type CreateCardBuilder struct {
	Brand           string
	Title           string
	Description     string
	VendorCode      string
	Dimensions      response.DimensionWrapper
	Sizes           []response.Size
	Characteristics []response.CharcWrapper
}

func NewCreateCardBuilder() *CreateCardBuilder {
	return &CreateCardBuilder{}
}

func (b *CreateCardBuilder) WithBrand(brand string) *CreateCardBuilder {
	b.Brand = brand
	return b
}
func (b *CreateCardBuilder) WithTitle(title string) *CreateCardBuilder {
	b.Title = title
	return b
}
func (b *CreateCardBuilder) WithDescription(description string) *CreateCardBuilder {
	b.Description = description
	return b
}
func (b *CreateCardBuilder) WithVendorCode(vendorCode string) *CreateCardBuilder {
	b.VendorCode = vendorCode
	return b
}
func (b *CreateCardBuilder) WithDimensions(dimensions response.Dimensions) *CreateCardBuilder {
	b.Dimensions = *dimensions.Unwrap()
	return b
}
func (b *CreateCardBuilder) WithDimensionWrapper(dimensions response.DimensionWrapper) *CreateCardBuilder {
	b.Dimensions = dimensions
	return b
}
func (b *CreateCardBuilder) WithSizes(sizes []response.Size) *CreateCardBuilder {
	b.Sizes = sizes
	return b
}
func (b *CreateCardBuilder) WithCharacteristics(charcs []response.CharcWrapper) *CreateCardBuilder {
	b.Characteristics = charcs
	return b
}

func (b *CreateCardBuilder) Build() (interface{}, error) {
	var card *request.CreateCardRequestData
	card = &request.CreateCardRequestData{
		Brand:           b.Brand,
		Title:           b.Title,
		Description:     b.Description,
		VendorCode:      b.VendorCode,
		Dimensions:      b.Dimensions,
		Sizes:           b.Sizes,
		Characteristics: b.Characteristics,
	}
	if err := card.Validate(); err != nil {
		return nil, err
	}
	return card, nil
}

func (b *CreateCardBuilder) Clear() {
	b.Brand = ""
	b.Title = ""
	b.Description = ""
	b.VendorCode = ""
	b.Dimensions = response.DimensionWrapper{}
	b.Sizes = []response.Size{}
	b.Characteristics = []response.CharcWrapper{}
}
