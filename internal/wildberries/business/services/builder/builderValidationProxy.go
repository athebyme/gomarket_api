package builder

import (
	response2 "gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type Proxy interface {
	Build() (interface{}, error)
	WithBrand(brand string) Proxy
	WithCharacteristics(characters []response2.CharcWrapper) Proxy
	WithSizes(sizes interface{}) Proxy
	WithDimensions(dimensions response2.DimensionWrapper) Proxy
	WithPrice(price int) Proxy
	WithVendorCode(code string) Proxy
	WithDescription(description string) Proxy
	WithTitle(title string) Proxy
}
