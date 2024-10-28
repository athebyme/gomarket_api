package business

type ProductService interface {
	GetAllGlobalIDs() ([]int, error)
}
