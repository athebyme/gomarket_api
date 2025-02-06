package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
)

type BrandRepository struct {
	db       *sql.DB
	prodRepo *ProductRepository
}

func NewBrandRepository(repo *ProductRepository) *BrandRepository {
	return &BrandRepository{prodRepo: repo, db: repo.db}
}

func (r *BrandRepository) GetProductBrandByID(productID int) (string, error) {
	query := `SELECT brand wholesaler.products WHERE global_id = $1`

	var brand string
	err := r.db.QueryRow(query, productID).Scan(&brand)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get product media: %w", err)
	}

	return brand, nil
}

func (r *BrandRepository) GetProductBrandByIDs(ids []int) (map[int]string, error) {
	query := `SELECT global_id, brand FROM wholesaler.products WHERE global_id = ANY($1)`

	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	brand := make(map[int]string)
	for rows.Next() {
		var globalId int
		var prodBrand string

		if err := rows.Scan(&globalId, &prodBrand); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		brand[globalId] = prodBrand
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return brand, nil
}

func (r *BrandRepository) GetProductsBrands() (map[int]string, error) {
	query := `SELECT global_id, brand FROM wholesaler.products`

	brand := make(map[int]string)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var globalId int
		var prodBrand string

		if err := rows.Scan(&globalId, &prodBrand); err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		brand[globalId] = prodBrand
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return brand, nil
}
