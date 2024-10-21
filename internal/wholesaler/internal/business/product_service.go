package business

import (
	"errors"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"log"
)

type ProductService struct {
	repo *storage.ProductRepository
}

func NewProductService(repo *storage.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(product *models.Product) error {
	if product.Model == "" || product.Category == "" {
		return errors.New("model and category are required")
	}

	log.Printf("Creating product: %s", product.Model)

	if err := s.repo.Insert(product); err != nil {
		return err
	}

	log.Printf("Product created with ID: %d", product.ID)
	return nil
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

func (s *ProductService) UpdateProduct(product *models.Product) error {
	if product.ID <= 0 {
		return errors.New("invalid product ID")
	}

	if err := s.repo.Update(product); err != nil {
		return err
	}

	log.Printf("Updated product with ID: %d", product.ID)
	return nil
}
func (s *ProductService) DeleteProduct(id int) error {
	if id <= 0 {
		return errors.New("invalid product ID")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	log.Printf("Deleted product with ID: %d", id)
	return nil
}
