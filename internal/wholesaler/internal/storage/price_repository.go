package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"log"
)

type PriceRepository struct {
	DB *sql.DB
}

func NewPriceRepository() (*PriceRepository, error) {
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

	log.Println("Successfully created wholesaler price repository")

	return &PriceRepository{DB: db}, nil
}

func (r *PriceRepository) GetPriceByProductID(id int) (*models.Price, error) {
	query := `
				SELECT global_id, price FROM wholesaler.price
				WHERE global_id = $1;
			 `
	var price models.Price
	err := r.DB.QueryRow(query, id).Scan(
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

func (r *PriceRepository) Close() error {
	return r.DB.Close()
}
