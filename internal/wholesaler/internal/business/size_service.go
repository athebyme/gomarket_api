package business

import (
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"gomarketplace_api/pkg/logger"
	"io"
)

type SizeService struct {
	repo   *repositories.SizeRepository
	logger logger.Logger
}

func NewSizeService(repo *repositories.SizeRepository, logWriter io.Writer) *SizeService {
	_log := logger.NewLogger(logWriter, "[SizeService]")
	_log.Log("SizeService successfully created.")
	return &SizeService{repo: repo, logger: _log}
}

func (s *SizeService) GetSizes() (map[int][]models.SizeWrapper, error) {
	sizeData, err := s.repo.GetSizes()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения размеров из репозитория: %w", err)
	}

	// Преобразуем в map[int][]models.SizeWrapper
	sizeMap := make(map[int][]models.SizeWrapper)
	for _, data := range sizeData {
		sizeMap[data.GlobalID] = append(sizeMap[data.GlobalID], data.Sizes...)
	}

	return sizeMap, nil
}

func (s *SizeService) GetSizesByIDs(ids []int) (map[int][]models.SizeWrapper, error) {
	sizeData, err := s.repo.GetSizesByIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения размеров из репозитория: %w", err)
	}

	// Преобразуем в map[int][]models.SizeWrapper
	sizeMap := make(map[int][]models.SizeWrapper)
	for _, data := range sizeData {
		sizeMap[data.GlobalID] = append(sizeMap[data.GlobalID], data.Sizes...)
	}

	return sizeMap, nil
}
