package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"log"
)

type PriceRepository struct {
	db      *sql.DB
	updater storage.Updater
}

func NewPriceRepository(db *sql.DB, updater storage.Updater) *PriceRepository {
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

func (r *PriceRepository) GetPrices() (map[int]int, error) {
	query := `
				SELECT global_id, price FROM wholesaler.price
			 `
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения prices: %w", err)
	}
	defer rows.Close()

	prices := make(map[int]int)
	for rows.Next() {
		var globalId int
		var price int
		if err := rows.Scan(&globalId, &price); err != nil {
			return nil, fmt.Errorf("ошибка сканирования prices: %w", err)
		}
		prices[globalId] = price
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return prices, nil
}

func (repo *PriceRepository) Update(args ...[]string) error {
	return repo.updater.Update(args...)
}

func (r *PriceRepository) Close() error {
	return r.db.Close()
}
