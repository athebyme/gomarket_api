package update

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"gomarketplace_api/internal/wildberries/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/business/models/dto/response"
	models "gomarketplace_api/internal/wildberries/business/models/get"
	"gomarketplace_api/internal/wildberries/business/services"
	"gomarketplace_api/internal/wildberries/business/services/update/operations"
	"gomarketplace_api/metrics"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	operation   operations.UpdateOperation
	uploadURL   string
	rateLimiter *rate.Limiter
	workerCount int
	metrics     *metrics.UpdateMetrics
	auth        services.AuthEngine
}

// NewUpdateService создает новый сервис обновления с указанными параметрами
func NewUpdateService(
	operation operations.UpdateOperation,
	uploadURL string,
	rateLimiter *rate.Limiter,
	workerCount int,
	auth services.AuthEngine,
) *Service {
	return &Service{
		operation:   operation,
		uploadURL:   uploadURL,
		rateLimiter: rateLimiter,
		workerCount: workerCount,
		metrics:     &metrics.UpdateMetrics{},
		auth:        auth,
	}
}

// Update запускает процесс обновления для номенклатур.
// nomenclatureChan – канал, в который поступают номенклатуры для обработки.
func (s *Service) Update(ctx context.Context, nomenclatureChan <-chan response.Nomenclature) (int, error) {
	processedItems := &sync.Map{}
	uploadChan := make(chan request.Model)

	var processWg sync.WaitGroup
	for i := 0; i < s.workerCount; i++ {
		processWg.Add(1)
		go func(workerID int) {
			defer processWg.Done()
			for nom := range nomenclatureChan {
				_, loaded := processedItems.LoadOrStore(nom.VendorCode, true)
				if loaded {
					continue
				}

				if !s.operation.Validate(nom) {
					log.Printf("Worker %d: номенклатура %s не прошла валидацию", workerID, nom.VendorCode)
					s.metrics.ErroredNomenclatures.Add(1)
					continue
				}

				model, err := s.operation.Process(ctx, nom)
				if err != nil {
					log.Printf("Worker %d: ошибка обработки %s: %s", workerID, nom.VendorCode, err)
					s.metrics.ErroredNomenclatures.Add(1)
					continue
				}
				uploadChan <- model
				s.metrics.ProcessedCount.Add(1)
			}
		}(i)
	}

	var uploadWg sync.WaitGroup
	uploadWg.Add(1)
	go func() {
		defer uploadWg.Done()
		s.uploadWorker(ctx, uploadChan)
	}()

	processWg.Wait()
	close(uploadChan)
	uploadWg.Wait()

	return int(s.metrics.UpdatedCount.Load()), nil
}

// uploadWorker обрабатывает загрузку данных
func (s *Service) uploadWorker(
	ctx context.Context,
	uploadChan <-chan request.Model) {
	s.UploadWorker(
		ctx,
		uploadChan,
		s.uploadURL,
		s.processAndUpload,
		&s.metrics.UpdatedCount,
	)
}

// UploadWorker отправляет модели на сервер с учетом rate limiter'а.
func (s *Service) UploadWorker(
	ctx context.Context,
	uploadChan <-chan request.Model,
	uploadURL string,
	processAndUploadFunc func(string, interface{}) (int, error),
	updatedCount *atomic.Int32,
) {
	for model := range uploadChan {
		if err := s.rateLimiter.Wait(ctx); err != nil {
			log.Printf("Limiter error: %s", err)
			continue
		}
		count, err := processAndUploadFunc(uploadURL, model)
		if err != nil {
			log.Printf("Error during upload: %s", err)
			continue
		}
		updatedCount.Add(int32(count))
	}
}

func (s *Service) processAndUpload(url string, data interface{}) (int, error) {
	bodyBytes, statusCode, err := s.uploadModels(url, data)
	if err == nil {
		return s.getDataLength(data), nil
	}

	if statusCode == http.StatusOK || bodyBytes == nil {
		return 0, err
	}

	log.Printf("Trying to fix (Status=%d)...", statusCode)

	bannedArticles, parseErr := s.extractBannedArticles(bodyBytes)
	if parseErr != nil {
		return 0, err
	}

	filteredModels := s.filterOutBannedModels(data, bannedArticles)
	if s.getDataLength(filteredModels) == 0 {
		return 0, err
	}

	return s.processAndUpload(url, filteredModels)
}

// extractBannedArticles извлекает список забаненных артикулов из ответа
func (s *Service) extractBannedArticles(bodyBytes []byte) ([]string, error) {
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &errorResponse); err != nil {
		return nil, err
	}

	additionalErrors, ok := errorResponse["additionalErrors"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("additionalErrors not found in response")
	}

	bannedArticlesStr, ok := additionalErrors["забаненные артикулы WB"].(string)
	if !ok {
		return nil, fmt.Errorf("забаненные артикулы WB not found in additionalErrors")
	}

	return strings.Split(bannedArticlesStr, ", "), nil
}

func (s *Service) filterOutBannedModels(data interface{}, bannedArticles []string) []interface{} {
	bannedSet := make(map[string]struct{}, len(bannedArticles))
	for _, article := range bannedArticles {
		bannedSet[article] = struct{}{}
		numberOfErroredNomenclatures.Add(1)
	}

	var filteredModels []interface{}

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		log.Println("Data is not a slice, returning empty filteredModels")
		return filteredModels
	}

	for i := 0; i < val.Len(); i++ {
		model := val.Index(i).Interface()
		var id string
		switch v := model.(type) {
		case *models.WildberriesCard:
			id = strconv.Itoa(v.NmID)
		case *request.MediaRequest:
			id = strconv.Itoa(v.NmId)
		}
		if _, isBanned := bannedSet[id]; !isBanned {
			filteredModels = append(filteredModels, model)
		}
	}

	return filteredModels
}

func (s *Service) uploadModels(url string, models interface{}) ([]byte, int, error) {
	log.Printf("Sending models to server...")

	requestBody, err := json.Marshal(models)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to marshal update request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, 500, fmt.Errorf("failed to create update request: %w", err)
	}

	requestBodySize := len(requestBody)
	requestBodySizeMB := float64(requestBodySize) / (1 << 20)
	log.Printf("Request Body Size: %d bytes (%.2f MB)\n", requestBodySize, requestBodySizeMB)

	req.Header.Set("Content-Type", "application/json")
	s.auth.SetApiKey(req)

	log.Printf("Sending request body: %v", string(requestBody))
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to upload models: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &errorResponse); err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed to unmarshal error response: %w", err)
		}

		errorDetails, err := json.MarshalIndent(errorResponse, "", "  ")
		if err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed to format error details: %w", err)
		}

		log.Printf("Upload failed with status: %d, error details: %s", resp.StatusCode, string(errorDetails))
		return bodyBytes, resp.StatusCode, fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	return bodyBytes, resp.StatusCode, nil
}

func (s *Service) getDataLength(data interface{}) int {
	switch v := data.(type) {
	case []interface{}:
		return len(v)
	case interface{}:
		return 1
	default:
		log.Println("Unknown data type, returning length as 0")
		return 0
	}
}

func (s *Service) Metrics() *metrics.UpdateMetrics {
	return s.metrics
}
