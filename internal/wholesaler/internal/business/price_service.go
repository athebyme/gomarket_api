package business

import (
	"errors"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"log"
)

type PriceService struct {
	repo *storage.PriceRepository
}

func NewPriceService(repo *storage.PriceRepository) *PriceService {
	return &PriceService{repo: repo}
}

func (s *PriceService) GetProductPriceByID(id int) (*models.Price, error) {
	if id <= 0 {
		return nil, errors.New("invalid product ID")
	}

	price, err := s.repo.GetPriceByProductID(id)
	if err != nil {
		return nil, err
	}

	if price == nil {
		return nil, errors.New("product not found")
	}

	log.Printf("Retrieved price with product with ID: %d", id)
	return price, nil
}
