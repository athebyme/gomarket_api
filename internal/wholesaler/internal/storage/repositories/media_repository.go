package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
)

type MediaRepository struct {
	prodRepo *ProductRepository
}

func NewMediaRepository(repo *ProductRepository) *MediaRepository {
	return &MediaRepository{prodRepo: repo}
}

func (r *MediaRepository) PopulateMediaTable() error {
	log.Printf("Checking for new media urls...")
	// Получение всех global_id из таблицы products
	rows, err := r.prodRepo.db.Query("SELECT global_id FROM wholesaler.products WHERE global_id NOT IN (SELECT global_id FROM wholesaler.media)")
	if err != nil {
		return fmt.Errorf("failed to fetch global_ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var globalID int
		if err := rows.Scan(&globalID); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		// Генерация медиа-URL для неконфиденциальных и конфиденциальных изображений
		mediaUrls, err := r.getMediaSourceByProductID(globalID, false)
		if err != nil {
			log.Printf("Failed to get media sources for global_id %d: %v", globalID, err)
			continue
		}

		mediaUrlsCensored, err := r.getMediaSourceByProductID(globalID, true)
		if err != nil {
			log.Printf("Failed to get censored media sources for global_id %d: %v", globalID, err)
			continue
		}

		// Вставка данных в таблицу media
		_, err = r.prodRepo.db.Exec("INSERT INTO wholesaler.media (global_id, images, images_censored) VALUES ($1, $2, $3)",
			globalID, pq.Array(mediaUrls), pq.Array(mediaUrlsCensored))
		if err != nil {
			log.Printf("Failed to insert media for global_id %d: %v", globalID, err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	log.Println("Media population completed successfully.")
	return nil
}

type ImageSize int

const (
	BigSize    ImageSize = 1200
	MediumSize ImageSize = 600
	SmallSize  ImageSize = 400
)

func (r *MediaRepository) GetMediaSourceByProductID(productID int, censored bool) ([]string, error) {
	// Изменим запрос для получения соответствующего массива ссылок
	query := `SELECT CASE WHEN $2 THEN images_censored ELSE images END FROM wholesaler.media WHERE global_id = $1`

	var mediaUrls []string
	err := r.prodRepo.db.QueryRow(query, productID, censored).Scan(
		pq.Array(&mediaUrls),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product media: %w", err)
	}

	return mediaUrls, nil
}

func (r *MediaRepository) GetMediaSourcesByProductIDs(productIDs []int, censored bool) (map[int][]string, error) {
	var err error
	mediaMap := make(map[int][]string)
	for _, productID := range productIDs {
		mediaMap[productID], err = r.GetMediaSourceByProductID(productID, censored)
		if err != nil {
			return nil, err
		}
	}
	return mediaMap, nil
}

func (r *MediaRepository) GetMediaSources(censored bool) (map[int][]string, error) {
	// Изменяем запрос в зависимости от значения параметра `censored`
	query := `SELECT global_id, CASE WHEN $1 THEN images_censored ELSE images END FROM wholesaler.media`

	rows, err := r.prodRepo.db.Query(query, censored)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса для получения media: %w", err)
	}
	defer rows.Close()

	mediaMap := make(map[int][]string)
	for rows.Next() {
		var globalId int
		var media []string
		if err := rows.Scan(&globalId, pq.Array(&media)); err != nil {
			return nil, fmt.Errorf("ошибка сканирования media: %w", err)
		}

		mediaMap[globalId] = media
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return mediaMap, nil
}

func (r *MediaRepository) getMediaSourceByProductID(productID int, censored bool) ([]string, error) {
	v, err := r.prodRepo.GetMediaSourceByProductID(productID, censored, BigSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error during getting media sources")
	}
	return v, nil
}
