package repositories

import (
	"fmt"
	"github.com/lib/pq"
	"strings"
)

type BarcodeRepository struct {
	prodRepo *ProductRepository
}

func NewBarcodeRepository(prodRepo *ProductRepository) *BarcodeRepository {
	return &BarcodeRepository{prodRepo}
}

func (r *BarcodeRepository) GetAllBarcodes() (map[int]interface{}, error) {
	query := `SELECT global_id, barcodes FROM wholesaler.products)`

	rows, err := r.prodRepo.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	barcodesMap := make(map[int]interface{})
	for rows.Next() {
		var globalId int
		var barcodes string
		if err := rows.Scan(&globalId, &barcodes); err != nil {
			return nil, fmt.Errorf("ошибка сканирования globalID: %w", err)
		}
		barcodesMap[globalId] = strings.Split(barcodes, "#")
	}

	return barcodesMap, nil
}

func (r *BarcodeRepository) GetBarcodesByProductIDs(ids []int) (map[int]interface{}, error) {
	query := `SELECT global_id, barcodes FROM wholesaler.products WHERE global_id=ANY($1)`

	rows, err := r.prodRepo.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения globalIDs: %w", err)
	}
	defer rows.Close()

	barcodesMap := make(map[int]interface{})
	for rows.Next() {
		var globalId int
		var barcodes string
		if err := rows.Scan(&globalId, &barcodes); err != nil {
			return nil, fmt.Errorf("ошибка сканирования globalID: %w", err)
		}
		barcodesMap[globalId] = strings.Split(barcodes, "#")
	}

	return barcodesMap, nil
}
