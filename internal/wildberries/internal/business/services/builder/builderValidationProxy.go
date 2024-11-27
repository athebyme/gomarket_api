package builder

import "gomarketplace_api/internal/wildberries/internal/business/models/dto/response"

type Proxy interface {
	Build() (interface{}, error)
	WithBrand(brand string) Proxy
	WithCharacteristics(characters []response.CharcWrapper) Proxy
	WithSizes(sizes interface{}) Proxy
	WithDimensions(dimensions response.DimensionWrapper) Proxy
	WithPrice(price int) Proxy
	WithVendorCode(code string) Proxy
	WithDescription(description string) Proxy
	WithTitle(title string) Proxy
}
