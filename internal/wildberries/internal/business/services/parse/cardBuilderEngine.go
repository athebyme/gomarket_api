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

func NewCardBuilderEngine(writer io.Writer, wildberriesValues values.WildberriesValues) *CardBuilderProxy {
	_log := logger.NewLogger(writer, "[CardBuilderProxy]")
	b := *builder.NewCreateCardBuilder()

	return &CardBuilderProxy{
		builder:           b,
		log:               _log,
		WildberriesValues: wildberriesValues,
	}
}

func (e *CardBuilderProxy) WithBrand(brand string) builder.Proxy {
	e.builder.WithBrand(brand)
	return e
}

func (e *CardBuilderProxy) WithTitle(title string) builder.Proxy {
	e.builder.WithTitle(title)
	return e
}

func (e *CardBuilderProxy) WithDescription(description string) builder.Proxy {
	e.builder.WithDescription(description)
	return e
}

func (e *CardBuilderProxy) WithVendorCode(code string) builder.Proxy {
	e.builder.WithVendorCode(code)
	return e
}

func (e *CardBuilderProxy) WithDimensions(dimensions response.DimensionWrapper) builder.Proxy {
	e.builder.WithDimensionWrapper(dimensions)
	return e
}

func (e *CardBuilderProxy) WithSizes(sizes interface{}) builder.Proxy {
	switch sizes.(type) {
	case response.SizeWrapper:
		e.builder.WithSizes(sizes.(response.SizeWrapper))
	case response.Size:
		size := sizes.(response.Size)
		e.builder.WithSizes(size.Wrap())
	default:
		e.log.FatalLog("Unexpected type of size")
	}
	return e
}

func (e *CardBuilderProxy) WithPrice(price int) builder.Proxy {
	// Проверяем текущий объект Sizes
	currentSizes := e.builder.Sizes

	// Создаём новый объект SizeWrapper, копируя существующие значения
	updatedSizes := response.SizeWrapper{
		TechSize: currentSizes.TechSize,
		WbSize:   currentSizes.WbSize,
		Skus:     currentSizes.Skus,
		Price:    price, // Обновляем только Price
	}

	// Устанавливаем обновлённый SizeWrapper
	e.builder.WithSizes(updatedSizes)

	return e
}

func (e *CardBuilderProxy) WithCharacteristics(characters []response.CharcWrapper) builder.Proxy {
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
