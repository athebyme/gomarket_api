package parse

import (
	"gomarketplace_api/config/values"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/internal/business/services/builder"
	"gomarketplace_api/pkg/logger"
	"io"
)

type CardBuilderProxy struct {
	builder builder.CreateCardBuilder
	log     logger.Logger
	values.WildberriesValues
}

func NewCardBuilderEngine(writer io.Writer) *CardBuilderProxy {
	_log := logger.NewLogger(writer, "[CardBuilderProxy]")
	b := *builder.NewCreateCardBuilder()

	return &CardBuilderProxy{
		builder: b,
		log:     _log,
	}
}

func (e *CardBuilderProxy) WithBrand(brand string) *CardBuilderProxy {
	e.builder.WithBrand(brand)
	return e
}

func (e *CardBuilderProxy) WithTitle(title string) *CardBuilderProxy {
	e.builder.WithTitle(title)
	return e
}

func (e *CardBuilderProxy) WithDescription(description string) *CardBuilderProxy {
	e.builder.WithDescription(description)
	return e
}

func (e *CardBuilderProxy) WithVendorCode(code string) *CardBuilderProxy {
	e.builder.WithVendorCode(code)
	return e
}

func (e *CardBuilderProxy) WithDimensions(dimensions response.DimensionWrapper) *CardBuilderProxy {
	e.builder.WithDimensionWrapper(dimensions)
	return e
}

func (e *CardBuilderProxy) WithSizes(sizes []response.Size) *CardBuilderProxy {
	e.builder.WithSizes(sizes)
	return e
}

func (e *CardBuilderProxy) WithCharacteristics(characters []response.CharcWrapper) *CardBuilderProxy {
	e.builder.WithCharacteristics(characters)
	return e
}

func (e *CardBuilderProxy) Build() (interface{}, error) {
	defer e.builder.Clear()

	e.WithDimensions(response.DimensionWrapper{
		Length: e.ensureDimension(e.builder.Dimensions.Length, e.WildberriesValues.PackageLength),
		Width:  e.ensureDimension(e.builder.Dimensions.Width, e.WildberriesValues.PackageWidth),
		Height: e.ensureDimension(e.builder.Dimensions.Height, e.WildberriesValues.PackageHeight),
	})

	return e.builder.Build()
}

func (e *CardBuilderProxy) ensureDimension(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}
