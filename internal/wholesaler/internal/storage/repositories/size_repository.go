package repositories

import (
	"database/sql"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/pkg/logger"
	"io"
)

type SizeRepository struct {
	db      *sql.DB
	updater storage.Updater
	logger  logger.Logger
}

func NewSizeRepository(db *sql.DB, updater storage.Updater, logWriter io.Writer) *SizeRepository {
	log := logger.NewLogger(logWriter, "[SizeRepository]")
	log.Log("SizeRepository successfully created.")
	return &SizeRepository{
		db:      db,
		updater: updater,
		logger:  log,
	}
}
func (r *SizeRepository) GetSizes() ([]models.SizeData, error) {
	query := `
		SELECT 
			sizes.global_id, 
			size_values.descriptor, 
			size_values.value_type, 
			size_values.value,
			size_values.unit
		FROM 
			wholesaler.sizes AS sizes
		JOIN 
			wholesaler.size_values AS size_values
		ON 
			sizes.size_id = size_values.size_id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	// Результат в виде списка SizeData
	var result []models.SizeData
	sizeMap := make(map[int]int) // Map для отслеживания индексов GlobalID в result

	for rows.Next() {
		var (
			globalID   int
			descriptor string
			sizeType   string
			value      float64
			unit       string
		)

		if err := rows.Scan(&globalID, &descriptor, &sizeType, &value, &unit); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}

		// Преобразуем данные в SizeWrapper
		sizeWrapper := models.SizeWrapper{
			Descriptor: models.SizeDescriptorEnum(descriptor),
			Type:       models.SizeTypeEnum(sizeType),
			Value:      value,
			Unit:       unit,
		}

		// Проверяем, есть ли уже GlobalID в result
		if idx, exists := sizeMap[globalID]; exists {
			// Добавляем новый SizeWrapper к существующему товару
			result[idx].Sizes = append(result[idx].Sizes, sizeWrapper)
		} else {
			// Создаем новый SizeData для GlobalID
			sizeMap[globalID] = len(result)
			result = append(result, models.SizeData{
				GlobalID: globalID,
				Sizes:    []models.SizeWrapper{sizeWrapper},
			})
		}
	}

	// Проверка на ошибки после итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return result, nil
}

func (r *SizeRepository) Update(args ...[]string) error {
	return r.updater.Update(args...)
}
func (r *SizeRepository) Close() error {
	return r.db.Close()
}
