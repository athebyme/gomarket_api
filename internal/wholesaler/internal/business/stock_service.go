package business

import (
	"errors"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"log"
)

type StockService struct {
	repo *storage.StocksRepository
}

func NewStockService(repo *storage.StocksRepository) *StockService {
	return &StockService{repo: repo}
}

func (s *StockService) GetProductStocksByID(id int) (*models.Stocks, error) {
	if id <= 0 {
		return nil, errors.New("invalid product ID")
	}

	stocks, err := s.repo.GetStocksByProductID(id)
	if err != nil {
		return nil, err
	}

	if stocks == nil {
		return nil, errors.New("product not found")
	}

	log.Printf("Retrieved (id=%d) stocks", id)
	return stocks, nil
}
