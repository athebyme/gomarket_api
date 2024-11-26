package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
)

type WbCategoriesRepository struct {
	db *sql.DB
}

func NewWbCategoriesRepository(db *sql.DB) *WbCategoriesRepository {
	return &WbCategoriesRepository{db: db}
}

func (r *WbCategoriesRepository) CategoryNameByID(id int) (*response.Category, error) {
	query := `
			SELECT category FROM wildberries.categories
			WHERE category_id = $1;
			 `
	var category response.Category
	err := r.db.QueryRow(query, id).Scan(
		&category.SubjectID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product price: %w", err)
	}
	return &category, nil
}
