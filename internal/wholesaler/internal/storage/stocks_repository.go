package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"log"
)

type StocksRepository struct {
	db      *sql.DB
	updater Updater
}

func NewStocksRepository(db *sql.DB, updater Updater) *StocksRepository {
	log.Println("Successfully connected to wholesaler stocks repository")
	return &StocksRepository{db: db, updater: updater}
}

func (r *StocksRepository) GetStocksByProductID(id int) (*models.Stocks, error) {
	query :=
		`
		SELECT global_id, main_articular, stocks FROM wholesaler.stocks
		WHERE global_id = $1
		`
	var stocks models.Stocks
	err := r.db.QueryRow(query, id).Scan(
		&stocks.ID, &stocks.MainArticular, &stocks.Stocks,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product price: %w", err)
	}

	return &stocks, nil
}

// Update args - аргументы для обновления. пока что поддерживается только ренейминг колонок
func (r *StocksRepository) Update(args ...[]string) error {
	return r.updater.Update(args...)
}

func (r *StocksRepository) Close() error {
	return r.db.Close()
}
