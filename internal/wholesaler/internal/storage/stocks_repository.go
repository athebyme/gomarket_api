package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"log"
)

type StocksRepository struct {
	DB *sql.DB
}

func NewStocksRepository() (*StocksRepository, error) {
	cfg := config.GetConfig()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to wholesaler stocks repository")

	return &StocksRepository{DB: db}, nil
}

func (r *StocksRepository) GetStocksByProductID(id int) (*models.Stocks, error) {
	query :=
		`
		SELECT global_id, main_articular, stocks FROM wholesaler.stocks
		WHERE global_id = $1
		`
	var stocks models.Stocks
	err := r.DB.QueryRow(query, id).Scan(
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

func (r *StocksRepository) Close() error {
	return r.DB.Close()
}
