package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"log"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository() (*ProductRepository, error) {
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

	return &ProductRepository{DB: db}, nil
}

func (r *ProductRepository) Insert(product *models.Product) error {
	query := `
		INSERT INTO wholesaler.products (
			global_id, model, appellation, category, brand, country, product_type, features, 
			sex, color, dimension, package, media, barcodes, material, package_battery
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, 
			$8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		RETURNING global_id`

	err := r.DB.QueryRow(
		query,
		product.Model, product.Appellation, product.Category, product.Brand, product.Country,
		product.ProductType, product.Features, product.Sex, product.Color,
		product.Dimension, product.Package, product.Media, product.Barcodes,
		product.Material, product.PackageBattery,
	).Scan(&product.ID)

	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}

	return nil
}

func (r *ProductRepository) GetProductByID(id int) (*models.Product, error) {
	query := `SELECT global_id, model, appellation, category, brand, country, product_type, features, 
				sex, color, dimension, package, media, barcodes, material, package_battery
			  FROM wholesaler.products WHERE global_id = $1`

	var product models.Product
	err := r.DB.QueryRow(query, id).Scan(
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

func (r *ProductRepository) Update(product *models.Product) error {
	query := `
		UPDATE wholesaler.products SET 
			model = $1, appellation = $2, category = $3, brand = $4, country = $5,
			product_type = $6, features = $7, sex = $8, color = $9, 
			dimension = $10, package = $11, media = $12, barcodes = $13, 
			material = $14, package_battery = $15
		WHERE global_id = $16`

	_, err := r.DB.Exec(
		query,
		product.Model, product.Appellation, product.Category, product.Brand, product.Country,
		product.ProductType, product.Features, product.Sex, product.Color,
		product.Dimension, product.Package, product.Media, product.Barcodes,
		product.Material, product.PackageBattery, product.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

func (r *ProductRepository) Delete(id int) error {
	query := `DELETE FROM wholesaler.products WHERE global_id = $1`

	_, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (r *ProductRepository) Close() error {
	return r.DB.Close()
}
