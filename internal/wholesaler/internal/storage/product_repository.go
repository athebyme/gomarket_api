package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"log"
)

type ProductRepository struct {
	db      *sql.DB
	updater Updater
}

func NewProductRepository(db *sql.DB, updater Updater) *ProductRepository {
	log.Printf("ProductRepositpory successfully created.")
	return &ProductRepository{
		db:      db,
		updater: updater}
}

func (r *ProductRepository) GetProductByID(id int) (*models.Product, error) {
	query := `SELECT global_id, model, appellation, category, brand, country, product_type, features, 
				sex, color, dimension, package, media, barcodes, material, package_battery
			  FROM wholesaler.products WHERE global_id = $1`

	var product models.Product
	err := r.db.QueryRow(query, id).Scan(
		&product.ID, &product.Model, &product.Appellation, &product.Category, &product.Brand, &product.Country,
		&product.ProductType, &product.Features, &product.Sex, &product.Color,
		&product.Dimension, &product.Package, &product.Media, &product.Barcodes,
		&product.Material, &product.PackageBattery,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) Update(args ...[]string) error {
	return r.updater.Update(args...)
}

func (r *ProductRepository) GetGlobalIDs() ([]int, error) {
	query := `SELECT global_id FROM wholesaler.products`

	rows, err := r.db.Query(query)
	if err != nil {
		return []int{}, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	var globalIDs []int
	// заполняем срез category_id из результата запроса
	for rows.Next() {
		var globalId int
		if err := rows.Scan(&globalId); err != nil {
			return []int{}, fmt.Errorf("ошибка сканирования globalID: %w", err)
		}
		globalIDs = append(globalIDs, globalId)
	}
	return globalIDs, nil
}

func (r *ProductRepository) GetAppellations() (map[int]string, error) {
	query := `SELECT global_id, appellation FROM wholesaler.products`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	appellations := make(map[int]string)
	for rows.Next() {
		var globalId int
		var appellation string
		if err := rows.Scan(&globalId, &appellation); err != nil {
			return nil, fmt.Errorf("ошибка сканирования globalID: %w", err)
		}
		appellations[globalId] = appellation
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return appellations, nil
}

func (r *ProductRepository) GetDescriptions() (map[int]string, error) {
	query := `SELECT global_id, product_description FROM wholesaler.descriptions`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	descriptions := make(map[int]string)
	for rows.Next() {
		var globalId int
		var description string
		if err := rows.Scan(&globalId, &description); err != nil {
			return nil, fmt.Errorf("ошибка сканирования globalID: %w", err)
		}
		descriptions[globalId] = description
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return descriptions, nil
}

func (r *ProductRepository) Close() error {
	return r.db.Close()
}
