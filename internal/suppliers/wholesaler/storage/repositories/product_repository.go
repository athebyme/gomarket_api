package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"gomarketplace_api/internal/suppliers/wholesaler/models"
	"log"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	log.Printf("ProductRepositpory successfully created.")
	return &ProductRepository{
		db: db}
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

func (r *ProductRepository) GetAppellations() (map[int]interface{}, error) {
	query := `SELECT global_id, appellation FROM wholesaler.products`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	appellations := make(map[int]interface{})
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

func (r *ProductRepository) GetAppellationsByIDs(ids []int) (map[int]interface{}, error) {
	query := `SELECT global_id, appellation FROM wholesaler.products WHERE global_id = ANY($1)`

	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	appellations := make(map[int]interface{})
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

func (r *ProductRepository) GetDescriptions(includeEmpty bool) (map[int]interface{}, error) {
	// Строим запрос на основе флага includeEmpty
	query := `
		SELECT 
			global_id, 
			product_description 
		FROM 
			wholesaler.descriptions
	`
	if !includeEmpty {
		query += ` WHERE product_description IS NOT NULL AND product_description != ''`
	}

	// Выполняем запрос
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения описаний: %w", err)
	}
	defer rows.Close()

	// Создаём map для результатов
	descriptions := make(map[int]interface{})
	for rows.Next() {
		var globalId int
		var description string
		if err := rows.Scan(&globalId, &description); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		descriptions[globalId] = description
	}

	// Проверяем ошибки при итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return descriptions, nil
}

func (r *ProductRepository) GetDescriptionsByIDs(ids []int, includeEmpty bool) (map[int]interface{}, error) {
	query := `SELECT global_id, product_description FROM wholesaler.descriptions WHERE global_id = ANY($1)`
	if !includeEmpty {
		query += ` AND product_description IS NOT NULL AND product_description != ''`
	}
	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	descriptions := make(map[int]interface{})
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
