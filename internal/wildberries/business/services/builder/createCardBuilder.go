package builder

import (
	"gomarketplace_api/internal/wildberries/business/models/dto/request"
	response2 "gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type CreateCardBuilder struct {
	Brand           string
	Title           string
	Description     string
	VendorCode      string
	Dimensions      response2.DimensionWrapper
	Sizes           []response2.SizeWrapper
	Characteristics []response2.CharcWrapper
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
func (b *CreateCardBuilder) WithDimensions(dimensions response2.Dimensions) *CreateCardBuilder {
	b.Dimensions = *dimensions.Unwrap()
	return b
}
func (b *CreateCardBuilder) WithDimensionWrapper(dimensions response2.DimensionWrapper) *CreateCardBuilder {
	b.Dimensions = dimensions
	return b
}
func (b *CreateCardBuilder) WithSizes(sizes response2.SizeWrapper) *CreateCardBuilder {
	b.Sizes = append(b.Sizes, sizes)
	return b
}
func (b *CreateCardBuilder) WithCharacteristics(charcs []response2.CharcWrapper) *CreateCardBuilder {
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
	if card.Characteristics == nil {
		card.Characteristics = []response2.CharcWrapper{}
	}
	for _, v := range card.Sizes {
		if v.Skus == nil {
			v.Skus = []string{}
		}
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
	b.Dimensions = response2.DimensionWrapper{}
	b.Sizes = []response2.SizeWrapper{}
	b.Characteristics = []response2.CharcWrapper{}
}
