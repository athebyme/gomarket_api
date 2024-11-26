package business

type PriceService interface {
	GetPriceById(id int) (float32, error)
	GetPrices(all bool) (interface{}, error)
}
