package business

import (
	"errors"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"gomarketplace_api/pkg/logger"
	"io"
)

type StockService struct {
	repo   *repositories.StocksRepository
	logger logger.Logger
}

func NewStockService(repo *repositories.StocksRepository, logWriter io.Writer) *StockService {
	_log := logger.NewLogger(logWriter, "[StockService]")
	_log.Log("StockService successfully created.")
	return &StockService{repo: repo, logger: _log}
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

	s.logger.Log("Retrieved (id=%d) stocks", id)
	return stocks, nil
}
