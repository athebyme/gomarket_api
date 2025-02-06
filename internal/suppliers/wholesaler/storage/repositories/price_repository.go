package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"gomarketplace_api/internal/suppliers/wholesaler/models"
)

type PriceRepository struct {
	db *sql.DB
}

func NewPriceRepository(db *sql.DB) *PriceRepository {
	return &PriceRepository{
		db: db,
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

func (r *PriceRepository) GetPricesById(ids []int) (map[int]interface{}, error) {
	query := `
				SELECT global_id, price FROM wholesaler.price
				WHERE global_id = ANY($1)
			 `
	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения prices: %w", err)
	}
	defer rows.Close()

	prices := make(map[int]interface{})
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

func (r *PriceRepository) GetPriceById(id int) (float32, error) {
	query := `
		SELECT price 
		FROM wholesaler.price
		WHERE global_id = $1
	`
	row := r.db.QueryRow(query, id)

	var price float32
	if err := row.Scan(&price); err != nil {
		if err == sql.ErrNoRows {
			return 0.0, fmt.Errorf("price not found for global_id: %d", id)
		}
		return 0.0, fmt.Errorf("error fetching price for global_id %d: %w", id, err)
	}

	return price, nil
}
