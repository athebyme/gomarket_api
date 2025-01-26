package parse

import (
	"gomarketplace_api/config/values"
	response2 "gomarketplace_api/internal/wildberries/business/models/dto/response"
	builder2 "gomarketplace_api/internal/wildberries/business/services/builder"
	"gomarketplace_api/pkg/logger"
	"io"
)

type CardBuilderProxy struct {
	builder builder2.CreateCardBuilder
	log     logger.Logger
	values.WildberriesValues
}

func NewCardBuilderEngine(writer io.Writer, wildberriesValues values.WildberriesValues) *CardBuilderProxy {
	_log := logger.NewLogger(writer, "[CardBuilderProxy]")
	b := *builder2.NewCreateCardBuilder()

	return &CardBuilderProxy{
		builder:           b,
		log:               _log,
		WildberriesValues: wildberriesValues,
	}
}

func (e *CardBuilderProxy) WithBrand(brand string) builder2.Proxy {
	e.builder.WithBrand(brand)
	return e
}

func (e *CardBuilderProxy) WithTitle(title string) builder2.Proxy {
	e.builder.WithTitle(title)
	return e
}

func (e *CardBuilderProxy) WithDescription(description string) builder2.Proxy {
	e.builder.WithDescription(description)
	return e
}

func (e *CardBuilderProxy) WithVendorCode(code string) builder2.Proxy {
	e.builder.WithVendorCode(code)
	return e
}

func (e *CardBuilderProxy) WithDimensions(dimensions response2.DimensionWrapper) builder2.Proxy {
	e.builder.WithDimensionWrapper(dimensions)
	return e
}

func (e *CardBuilderProxy) WithSizes(sizes interface{}) builder2.Proxy {
	switch sizes.(type) {
	case response2.SizeWrapper:
		e.builder.WithSizes(sizes.(response2.SizeWrapper))
	case response2.Size:
		size := sizes.(response2.Size)
		e.builder.WithSizes(size.Wrap())
	default:
		e.log.FatalLog("Unexpected type of size")
	}
	return e
}

func (e *CardBuilderProxy) WithPrice(price int) builder2.Proxy {
	// Проверяем текущий объект Sizes
	var currentSizes response2.SizeWrapper
	if e.builder.Sizes == nil || len(e.builder.Sizes) == 0 {
		currentSizes = response2.SizeWrapper{}
	} else {
		currentSizes = e.builder.Sizes[0]
	}

	// Создаём новый объект SizeWrapper, копируя существующие значения
	updatedSizes := response2.SizeWrapper{
		TechSize: currentSizes.TechSize,
		WbSize:   currentSizes.WbSize,
		Skus:     currentSizes.Skus,
		Price:    price, // Обновляем только Price
	}

	// Устанавливаем обновлённый SizeWrapper
	e.builder.WithSizes(updatedSizes)

	return e
}

func (e *CardBuilderProxy) WithCharacteristics(characters []response2.CharcWrapper) builder2.Proxy {
	e.builder.WithCharacteristics(characters)
	return e
}

func (e *CardBuilderProxy) Build() (interface{}, error) {
	defer e.builder.Clear()

	e.WithDimensions(response2.DimensionWrapper{
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
