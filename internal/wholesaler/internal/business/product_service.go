package business

import (
	"errors"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"log"
)

type ProductService struct {
	repo *repositories.ProductRepository
}

func NewProductService(repo *repositories.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetProductByID(id int) (*models.Product, error) {
	if id <= 0 {
		return nil, errors.New("invalid product ID")
	}

	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, errors.New("product not found")
	}

	log.Printf("Retrieved product with ID: %d", id)
	return product, nil
}

func (s *ProductService) GetProductsByIDs(ids []int) ([]*models.Product, error) {
	var products []*models.Product

	for _, id := range ids {
		if id <= 0 {
			log.Printf("Skipping invalid product ID: %d", id)
			continue
		}

		product, err := s.repo.GetProductByID(id)
		if err != nil {
			log.Printf("Error retrieving product with ID %d: %v", id, err)
			continue
		}

		if product == nil {
			log.Printf("Product with ID %d not found", id)
			continue
		}

		products = append(products, product)
		log.Printf("Retrieved product with ID: %d", id)
	}

	return products, nil
}

func (s *ProductService) GetAllGlobalIDs() ([]int, error) {
	res, err := s.repo.GetGlobalIDs()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ProductService) GetAllAppellations() (map[int]interface{}, error) {
	res, err := s.repo.GetAppellations()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ProductService) GetAllAppellationsByIDs(ids []int) (map[int]interface{}, error) {
	res, err := s.repo.GetAppellationsByIDs(ids)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ProductService) GetAllDescriptions(includeEmpty bool) (map[int]interface{}, error) {
	res, err := s.repo.GetDescriptions(includeEmpty)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ProductService) GetAllDescriptionsByIDs(ids []int, includeEmpty bool) (map[int]interface{}, error) {
	res, err := s.repo.GetDescriptionsByIDs(ids, includeEmpty)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ProductService) GetAllMediaSources(censored bool, imgSize repositories.ImageSize) (map[int][]string, error) {
	res, err := s.repo.GetMediaSources(censored, imgSize)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ProductService) GetAllMediaSourcesByProductIDs(globalIds []int, censored bool, imgSize repositories.ImageSize) (map[int][]string, error) {
	res, err := s.repo.GetMediaSourcesByProductIDs(globalIds, censored, imgSize)
	if err != nil {
		return nil, err
	}
	return res, nil
}
