package update

import (
	"context"
)

// Провайдеры обязаны возвращать map[id]<T>, где T - тип данных требуемого параметра

type DataProvider interface {
	GetIds(ctx context.Context) (map[int]struct{}, error)
	GetBrands(ctx context.Context) (map[int]string, error)
	GetAppellations(ctx context.Context) (map[int]string, error)
	GetDescriptions(ctx context.Context) (map[int]string, error)
	GetPrices(ctx context.Context) (map[int]float64, error)
	GetBarcodes(ctx context.Context) (map[int][]string, error)
	GetSizes(ctx context.Context) (map[int]interface{}, error)
}

type IDsProvider interface {
	// GetIds возвращает не слайс, а map[int] для константного доступа к id
	GetIds(ctx context.Context) (map[int]struct{}, error)
}

type BrandProvider interface {
	GetBrands(ctx context.Context) (map[int]string, error)
}

type AppellationProvider interface {
	GetAppellations(ctx context.Context) (map[int]string, error)
}

type DescriptionProvider interface {
	GetDescriptions(ctx context.Context) (map[int]string, error)
}

type PriceProvider interface {
	GetPrices(ctx context.Context) (map[int]float64, error)
}

type BarcodeProvider interface {
	GetBarcodes(ctx context.Context) (map[int][]string, error)
}

type SizesProvider interface {
	GetSizes(ctx context.Context) (map[int]interface{}, error)
}
