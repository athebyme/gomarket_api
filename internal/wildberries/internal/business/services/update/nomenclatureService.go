package update

import (
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	"gomarketplace_api/internal/wildberries/internal/storage"
)

type NomenclatureService struct {
	nmEngine get.NomenclatureEngine
	repo     storage.NomenclatureRepository
}

func NewNomenclatureService(nmEngine get.NomenclatureEngine, repo storage.NomenclatureRepository) *NomenclatureService {
	return &NomenclatureService{nmEngine: nmEngine, repo: repo}
}

func (nms *NomenclatureService) GetSetOfUncreatedItemsWithCategories(accuracy float32, inStock bool, categoryID int) (map[int]interface{}, error) {
	return nms.repo.GetSetOfUncreatedItemsWithCategories(accuracy, inStock, categoryID)
}
