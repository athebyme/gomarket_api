package business

type PriceService interface {
	GetPriceById(id int) (int, error)
	GetPricesById(ids []int) (map[int]interface{}, error)
	GetPrices(all bool) (interface{}, error)
}
