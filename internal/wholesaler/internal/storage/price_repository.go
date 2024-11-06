package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"log"
)

type PriceRepository struct {
	db      *sql.DB
	updater Updater
}

func NewPriceRepository(db *sql.DB, updater Updater) *PriceRepository {
	log.Println("Successfully created wholesaler price repository")
	return &PriceRepository{
		db:      db,
		updater: updater,
	}
}

func (r *PriceRepository) GetPriceByProductID(id int) (*models.Price, error) {
	query := `
				SELECT global_id, price FROM wholesaler.price
				WHERE global_id = $1;
			 `
	var price models.Price
	err := r.db.QueryRow(query, id).Scan(
		&price.ID, &price.Price,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product price: %w", err)
	}
	return &price, nil
}

func (repo *PriceRepository) Update(args ...[]string) error {
	return repo.updater.Update(args...)
}

func (r *PriceRepository) Close() error {
	return r.db.Close()
}
