package update

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"gomarketplace_api/config/values"
	"gomarketplace_api/internal/wholesaler/pkg/clients"
	"gomarketplace_api/internal/wholesaler/pkg/requests"
	request2 "gomarketplace_api/internal/wildberries/business/models/dto/request"
	response2 "gomarketplace_api/internal/wildberries/business/models/dto/response"
	models "gomarketplace_api/internal/wildberries/business/models/get"
	"gomarketplace_api/internal/wildberries/business/services"
	"gomarketplace_api/internal/wildberries/business/services/builder"
	"gomarketplace_api/internal/wildberries/business/services/get"
	"gomarketplace_api/internal/wildberries/business/services/parse"
	clients2 "gomarketplace_api/internal/wildberries/pkg/clients"
	"gomarketplace_api/pkg/business/service"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	uploadBatchSize  = 2000
	maxBatchSize     = 1 << 20 // 1 MB
	goroutineCount   = 5
	requestRateLimit = 70 // requests per minute
	uploadRateLimit  = 10
	maxTitleLength   = 60
	maxDescLength    = 2000
)

type CardProcessor struct {
	nomenclature response2.Nomenclature
	globalID     int
	appellation  string
	description  string
}

type CardUpdateService struct {
	nomenclatureService get.SearchEngine
	wsclient            *clients2.WServiceClient
	textService         service.ITextService
	cardBuilder         *builder.CardBuilder
	brandService        parse.BrandService
	defaultValues       values.WildberriesValues
	metrics             *updateMetrics
	services.AuthEngine
}

type updateMetrics struct {
	updatedCount                 atomic.Int32
	numberOfErroredNomenclatures atomic.Int32
	goroutinesNmsCount           atomic.Int32
}

type batchProcessor struct {
	batch    []request2.Model
	size     int
	mu       sync.Mutex
	uploadCh chan<- []request2.Model
}

func (bp *batchProcessor) add(card *models.WildberriesCard) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.batch = append(bp.batch, card)
	bp.size += len([]byte(card.Title)) + len([]byte(card.Description))

	if len(bp.batch) >= uploadBatchSize || bp.size >= maxBatchSize {
		bp.flush()
	}
}

func (bp *batchProcessor) flush() {
	if len(bp.batch) > 0 {
		batchToSend := make([]request2.Model, len(bp.batch))
		copy(batchToSend, bp.batch)
		bp.uploadCh <- batchToSend
		bp.batch = nil
		bp.size = 0
	}
}

func NewCardUpdateService(nservice *get.SearchEngine, textService service.ITextService, wsClientUrl string, auth services.AuthEngine, writer io.Writer, brandService parse.BrandService, wbDefaultValues values.WildberriesValues) *CardUpdateService {
	return &CardUpdateService{
		nomenclatureService: *nservice,
		wsclient:            clients2.NewWServiceClient(wsClientUrl, writer),
		textService:         textService,
		brandService:        brandService,
		AuthEngine:          auth,
		cardBuilder:         builder.NewCardBuilder(textService),
		defaultValues:       wbDefaultValues,
	}
}

const updateCardsUrl = "https://content-api.wildberries.ru/content/v2/cards/update"

var numberOfErroredNomenclatures atomic.Int32
var updatedCount atomic.Int32

func (cu *CardUpdateService) UpdateCardNaming(ctx context.Context, settings request2.Settings) (int, error) {
	// Подготовка данных
	appellationsMap, descriptionsMap, err := cu.fetchRequiredData(ctx)
	if err != nil {
		return 0, err
	}

	// Инициализация каналов и лимитеров
	nomenclatureCh := make(chan response2.Nomenclature)
	uploadCh := make(chan []request2.Model)
	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/requestRateLimit), requestRateLimit)
	uploadLimiter := rate.NewLimiter(rate.Every(time.Minute/uploadRateLimit), uploadRateLimit)

	// Создание процессора пакетов
	batchProc := &batchProcessor{
		uploadCh: uploadCh,
	}

	// Настройка групп ожидания
	var processWg sync.WaitGroup
	var uploadWg sync.WaitGroup
	processedItems := &sync.Map{}

	// Запуск получения номенклатур
	go cu.fetchNomenclatures(ctx, settings, nomenclatureCh, responseLimiter)

	// Запуск обработчиков
	for i := 0; i < goroutineCount; i++ {
		processWg.Add(1)
		go cu.processNomenclatures(
			ctx,
			i,
			nomenclatureCh,
			batchProc,
			processedItems,
			appellationsMap,
			descriptionsMap,
			&processWg,
		)
	}

	// Запуск загрузчика
	uploadWg.Add(1)
	go cu.uploadWorker(ctx, uploadCh, uploadLimiter, &uploadWg)

	// Ожидание завершения и обработка оставшихся данных
	processWg.Wait()
	close(uploadCh)
	uploadWg.Wait()

	cu.logResults()
	return int(cu.metrics.updatedCount.Load()), nil
}

// считываем канал номенклатур, валидируем и отдаем
func (cu *CardUpdateService) processNomenclatures(
	ctx context.Context,
	workerID int,
	nomenclatureCh <-chan response2.Nomenclature,
	batchProc *batchProcessor,
	processedItems *sync.Map,
	appellationsMap, descriptionsMap map[int]interface{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for nomenclature := range nomenclatureCh {
		cu.metrics.goroutinesNmsCount.Add(1)

		processor := &CardProcessor{nomenclature: nomenclature}
		if !cu.validateAndPrepareProcessor(processor, processedItems, appellationsMap, descriptionsMap) {
			continue
		}

		card := cu.cardBuilder.
			FromNomenclature(nomenclature).
			WithUpdatedTitle(processor.appellation, maxTitleLength)

		if processor.description != "" {
			card.WithDescription(processor.description, maxDescLength)
		} else {
			card.WithFallbackDescription(processor.appellation, maxDescLength)
		}

		batchProc.add(card.Build())
	}
}

func (cu *CardUpdateService) validateAndPrepareProcessor(
	processor *CardProcessor,
	processedItems *sync.Map,
	appellationsMap, descriptionsMap map[int]interface{},
) bool {
	// Проверка на дубликаты
	if _, loaded := processedItems.LoadOrStore(processor.nomenclature.VendorCode, true); loaded {
		return false
	}

	// Получение и валидация globalID
	globalID, err := processor.nomenclature.GlobalID()
	if err != nil || globalID == 0 {
		cu.metrics.numberOfErroredNomenclatures.Add(1)
		return false
	}

	// Проверка наличия appellation
	appellation, ok := appellationsMap[globalID]
	if !ok {
		cu.metrics.numberOfErroredNomenclatures.Add(1)
		return false
	}

	processor.globalID = globalID
	processor.appellation = appellation.(string)
	if description, ok := descriptionsMap[globalID]; ok {
		processor.description = description.(string)
	}

	return true
}

func (cu *CardUpdateService) fetchRequiredData(ctx context.Context) (map[int]interface{}, map[int]interface{}, error) {
	log.Println("Fetching filterAppellations...")
	appellationsMap, err := cu.wsclient.FetchAppellations(requests.AppellationsRequest{
		FilterRequest: requests.FilterRequest{ProductIDs: []int{}},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching filterAppellations: %w", err)
	}

	log.Println("Fetching descriptions...")
	descriptionsMap, err := cu.wsclient.FetchDescriptions(requests.AppellationsRequest{
		FilterRequest: requests.FilterRequest{ProductIDs: []int{}},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching descriptions: %w", err)
	}

	return appellationsMap, descriptionsMap, nil
}

func (cu *CardUpdateService) fetchNomenclatures(
	ctx context.Context,
	settings request2.Settings,
	nomenclatureCh chan<- response2.Nomenclature,
	responseLimiter *rate.Limiter,
) {
	defer close(nomenclatureCh)

	err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(
		ctx,
		settings,
		"",
		nomenclatureCh,
		responseLimiter,
	)
	if err != nil {
		log.Printf("Error fetching nomenclatures concurrently: %s", err)
		return
	}
}

func (cu *CardUpdateService) uploadWorker(
	ctx context.Context,
	uploadCh <-chan []request2.Model,
	uploadLimiter *rate.Limiter,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for batch := range uploadCh {
		select {
		case <-ctx.Done():
			return
		default:
			log.Println("Uploading batch of cards...")

			if err := uploadLimiter.Wait(ctx); err != nil {
				log.Printf("Error waiting for rate limiter: %s", err)
				return
			}

			cards, err := cu.processAndUpload(updateCardsUrl, batch)
			if err != nil {
				log.Printf("Error during uploading: %s", err)
				continue // Продолжаем с следующим батчем вместо полной остановки
			}

			cu.metrics.updatedCount.Add(int32(cards))
		}
	}
}

func (cu *CardUpdateService) logResults() {
	log.Printf("Goroutines fetchers got (%d) nomenclatures",
		cu.metrics.goroutinesNmsCount.Load())
	log.Printf("Update completed, total updated count: %d. Unfetched count: %d",
		cu.metrics.updatedCount.Load(),
		cu.metrics.numberOfErroredNomenclatures.Load())
}

func (cu *CardUpdateService) UpdateCardPackages(settings request2.Settings) (int, error) {
	const UPLOAD_SIZE = 2000
	const MaxBatchSize = 1 << 20 // 1 MB
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 70 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 10
	var currentBatch []request2.Model
	var currentBatchSize int
	var gotData []response2.Nomenclature
	var goroutinesNmsCount atomic.Int32

	var processWg sync.WaitGroup
	var uploadWg sync.WaitGroup
	var mu sync.Mutex
	var processedItems sync.Map
	nomenclatureChan := make(chan response2.Nomenclature)
	uploadChan := make(chan []request2.Model) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)

	log.Println("Fetching and sending nomenclatures to the channel...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(ctx, settings, "", nomenclatureChan, responseLimiter); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
	}()

	// Запуск горутин для обработки номенклатур
	log.Println("Starting goroutines for processing nomenclatures...")
	for i := 0; i < GOROUTINE_COUNT; i++ {
		processWg.Add(1)
		go func(i int) {
			defer processWg.Done()
			for nomenclature := range nomenclatureChan {
				goroutinesNmsCount.Add(1)

				var wbCard models.WildberriesCard
				wbCard = *wbCard.FromNomenclature(nomenclature)
				_, loaded := processedItems.LoadOrStore(wbCard.VendorCode, true)
				if loaded {
					continue // Если запись уже была обработана, пропускаем её
				}

				globalId, err := nomenclature.GlobalID()
				if err != nil || globalId == 0 {
					log.Printf("(G%d) (globalID=%s) parse error (not SPB aricular)", i, nomenclature.VendorCode)
					numberOfErroredNomenclatures.Add(1)
					continue
				}

				wbCard.Dimensions = response2.DimensionWrapper{
					Length: cu.defaultValues.PackageLength,
					Width:  cu.defaultValues.PackageWidth,
					Height: cu.defaultValues.PackageHeight,
				}

				mu.Lock()
				gotData = append(gotData, nomenclature)
				currentBatch = append(currentBatch, &wbCard)
				currentBatchSize += len([]byte(wbCard.Title)) + len([]byte(wbCard.Description))
				if len(currentBatch) >= UPLOAD_SIZE || currentBatchSize >= MaxBatchSize {
					batchToSend := currentBatch
					currentBatch = nil
					currentBatchSize = 0
					uploadChan <- batchToSend
				}
				mu.Unlock()
			}
		}(i)
	}

	// Горутин для отправки данных
	uploadWg.Add(1)
	go func() {
		defer uploadWg.Done()
		for batch := range uploadChan {
			log.Println("Uploading batch of cards...")
			// Лимитирование запросов на загрузку
			if err := uploadToServerLimiter.Wait(context.Background()); err != nil {
				log.Printf("Error waiting for rate limiter: %s", err)
				return
			}

			cards, err := cu.processAndUpload(updateCardsUrl, batch)
			if err != nil {
				log.Printf("Error during uploading %s", err)
				return
			}
			updatedCount.Add(int32(cards))
		}
	}()

	processWg.Wait()

	// Process remaining data
	log.Println("Processing any remaining data...")
	if len(currentBatch) > 0 {
		log.Println("Uploading remaining batch of cards...")
		uploadChan <- currentBatch
		currentBatch = nil
	}
	close(uploadChan)
	uploadWg.Wait()

	log.Printf("Goroutines fetchers got (%d) nomenclautres", goroutinesNmsCount.Load())
	log.Printf("Update completed, total updated count: %d. Unfetched count : %d", updatedCount.Load(), numberOfErroredNomenclatures.Load())
	return int(updatedCount.Load()), nil
}

const updateCardsMediaUrl = "https://content-api.wildberries.ru/content/v3/media/save"

func (cu *CardUpdateService) UpdateCardMedia(settings request2.Settings) (int, error) {
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 60 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 60

	var updatedCount = 0
	var goroutinesNmsCount atomic.Int32

	var processWg sync.WaitGroup
	var uploadWg sync.WaitGroup
	var mu sync.Mutex
	var processedItems sync.Map

	log.Println("Fetching media urls...")

	mediaRequest := clients.ImageRequest{
		Censored: false,
	}
	mediaMap, err := cu.wsclient.FetchImages(mediaRequest)
	if err != nil {
		return 0, fmt.Errorf("error fetching media urls: %w", err)
	}

	nomenclatureChan := make(chan response2.Nomenclature)
	uploadChan := make(chan request2.Model) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Println("Fetching and sending nomenclatures to the channel...")
	go func() {
		if err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(ctx, settings, "", nomenclatureChan, responseLimiter); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
	}()

	// Запуск горутин для обработки номенклатур
	log.Println("Starting goroutines for processing nomenclatures...")
	for i := 0; i < GOROUTINE_COUNT; i++ {
		processWg.Add(1)
		go func(i int) {
			defer processWg.Done()
			for nomenclature := range nomenclatureChan {
				goroutinesNmsCount.Add(1)

				_, loaded := processedItems.LoadOrStore(nomenclature.VendorCode, true)
				if loaded {
					continue // Если запись уже была обработана, пропускаем её
				}

				globalId, err := nomenclature.GlobalID()
				if err != nil || globalId == 0 {
					log.Printf("(G%d) (globalID=%s) parse error (not SPB aricular)", i, nomenclature.VendorCode)
					numberOfErroredNomenclatures.Add(1)
					continue
				}

				if len(nomenclature.Photos) != 1 {
					continue
				}

				var urls []string
				var ok bool
				if urls, ok = mediaMap[globalId]; !ok && len(urls) <= 0 {
					log.Printf("(G%d) (globalID=%s) not found urls !", i, nomenclature.VendorCode)
					numberOfErroredNomenclatures.Add(1)
					continue
				}

				// +1 photo (copy) if only 1 exits (to improve card quality)
				if len(urls) == 1 {
					urls = append(urls, urls[0])
				}

				if len(urls) == 1 {
					urls = append(urls, urls[0])
				}
				mediaRequest := request2.NewMediaRequest(nomenclature.NmID, urls)

				if err := responseLimiter.Wait(context.Background()); err != nil {
					log.Printf("Rate limiter error: %s", err)
					return
				}
				if err != nil {
					log.Printf("Error fetching nomenclature %d: %s", i, err)
					return
				}
				mu.Lock()
				uploadChan <- mediaRequest
				mu.Unlock()
			}
		}(i)
	}

	// Горутин для отправки данных
	uploadWg.Add(1)
	go func() {
		defer uploadWg.Done()
		for batch := range uploadChan {
			log.Println("Uploading batch of media...")
			// Лимитирование запросов на загрузку
			if err := uploadToServerLimiter.Wait(context.Background()); err != nil {
				log.Printf("Error waiting for rate limiter: %s", err)
				return
			}

			media, err := cu.processAndUpload(updateCardsMediaUrl, batch)
			if err != nil {
				log.Printf("Error during uploading %s", err)
				continue
			}
			updatedCount += media
		}
	}()

	processWg.Wait()
	close(uploadChan)
	uploadWg.Wait()

	log.Printf("Goroutines fetchers got (%d) nomenclautres", goroutinesNmsCount.Load())
	log.Printf("Media update completed, total updated count: %d. Unfetched count : %d", updatedCount, numberOfErroredNomenclatures.Load())
	return updatedCount, nil
}

func (cu *CardUpdateService) UpdateCardBrand(settings request2.Settings) (int, error) {
	const UPLOAD_SIZE = 2000
	const MaxBatchSize = 1 << 20 // 1 MB
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 70 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 10
	var currentBatch []request2.Model
	var currentBatchSize int
	var gotData []response2.Nomenclature
	var goroutinesNmsCount atomic.Int32

	log.Println("Fetching Brands list...")
	brandsMap, err := cu.wsclient.FetchBrands(requests.BrandRequest{FilterRequest: requests.FilterRequest{ProductIDs: []int{}}})
	if err != nil {
		return 0, fmt.Errorf("error fetching descriptions: %w", err)
	}

	var processWg sync.WaitGroup
	var uploadWg sync.WaitGroup
	var mu sync.Mutex
	var processedItems sync.Map
	nomenclatureChan := make(chan response2.Nomenclature)
	uploadChan := make(chan []request2.Model) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)

	log.Println("Fetching and sending nomenclatures to the channel...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChannel(ctx, settings, "", nomenclatureChan, responseLimiter); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
	}()

	// Запуск горутин для обработки номенклатур
	log.Println("Starting goroutines for processing nomenclatures...")
	for i := 0; i < GOROUTINE_COUNT; i++ {
		processWg.Add(1)
		go func(i int) {
			defer processWg.Done()
			for nomenclature := range nomenclatureChan {
				goroutinesNmsCount.Add(1)

				var wbCard models.WildberriesCard
				wbCard = *wbCard.FromNomenclature(nomenclature)
				_, loaded := processedItems.LoadOrStore(wbCard.VendorCode, true)
				if loaded {
					continue // Если запись уже была обработана, пропускаем её
				}

				globalId, err := nomenclature.GlobalID()
				if err != nil || globalId == 0 {
					numberOfErroredNomenclatures.Add(1)
					continue
				}

				switch brand := brandsMap[globalId].(type) {
				case string:
					if cu.brandService.IsBanned(brand) || brand == "" {
						numberOfErroredNomenclatures.Add(1)
						continue
					}
					wbCard.Brand = brand
				default:
					numberOfErroredNomenclatures.Add(1)
				}

				mu.Lock()
				gotData = append(gotData, nomenclature)
				currentBatch = append(currentBatch, &wbCard)
				currentBatchSize += len([]byte(wbCard.Title)) + len([]byte(wbCard.Description))
				if len(currentBatch) >= UPLOAD_SIZE || currentBatchSize >= MaxBatchSize {
					batchToSend := currentBatch
					currentBatch = nil
					currentBatchSize = 0
					uploadChan <- batchToSend
				}
				mu.Unlock()
			}
		}(i)
	}

	// Горутин для отправки данных
	uploadWg.Add(1)
	go func() {
		defer uploadWg.Done()
		for batch := range uploadChan {
			log.Println("Uploading batch of cards...")
			// Лимитирование запросов на загрузку
			if err := uploadToServerLimiter.Wait(context.Background()); err != nil {
				log.Printf("Error waiting for rate limiter: %s", err)
				return
			}

			cards, err := cu.processAndUpload(updateCardsUrl, batch)
			if err != nil {
				log.Printf("Error during uploading %s", err)
				return
			}
			updatedCount.Add(int32(cards))
		}
	}()

	processWg.Wait()

	// Process remaining data
	log.Println("Processing any remaining data...")
	if len(currentBatch) > 0 {
		log.Println("Uploading remaining batch of cards...")
		uploadChan <- currentBatch
		currentBatch = nil
	}
	close(uploadChan)
	uploadWg.Wait()

	log.Printf("Goroutines fetchers got (%d) nomenclautres", goroutinesNmsCount.Load())
	log.Printf("Update completed, total updated count: %d. Unfetched count : %d", updatedCount.Load(), numberOfErroredNomenclatures.Load())
	return int(updatedCount.Load()), nil
}

func (cu *CardUpdateService) UpdateDBNomenclatures(settings request2.Settings, locale string) (int, error) {
	return cu.nomenclatureService.UploadToDb(settings, locale)
}

func (cu *CardUpdateService) CheckSearchEngine(settings request2.Settings, locale string) (int, error) {
	return cu.nomenclatureService.CheckTotalNmCount(settings, locale)
}

func (cu *CardUpdateService) processAndUpload(url string, data interface{}) (int, error) {
	bodyBytes, statusCode, err := cu.uploadModels(url, data)
	if err != nil {
		if statusCode != http.StatusOK && bodyBytes != nil {
			var errorResponse map[string]interface{}
			log.Printf("Trying to fix (Status=%d)...", statusCode)
			if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil {
				if additionalErrors, ok := errorResponse["additionalErrors"].(map[string]interface{}); ok {
					if bannedArticles, ok := additionalErrors["Забаненные артикулы WB"].(string); ok {
						bannedArticlesSlice := strings.Split(bannedArticles, ", ")
						filteredModels := cu.filterOutBannedModels(data, bannedArticlesSlice)
						if len(filteredModels) > 0 {
							dataLen, err := cu.processAndUpload(url, filteredModels)
							if err != nil {
								return 0, err
							}
							return dataLen, nil
						}
					}
				}
			}
		}
	}

	// Используем универсальную функцию для получения длины данных
	return cu.getDataLength(data), nil
}

func (cu *CardUpdateService) getDataLength(data interface{}) int {
	// Проверяем, является ли data срезом
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

func (cu *CardUpdateService) uploadModels(url string, models interface{}) ([]byte, int, error) {
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
	cu.SetApiKey(req)

	log.Printf("Sending request body: %v", string(requestBody))
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to upload models: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
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

func (cu *CardUpdateService) filterOutBannedModels(data interface{}, bannedArticles []string) []interface{} {
	bannedSet := make(map[string]struct{}, len(bannedArticles))
	for _, article := range bannedArticles {
		bannedSet[article] = struct{}{}
		numberOfErroredNomenclatures.Add(1)
	}

	var filteredModels []interface{}

	// Используем reflection для обработки срезов
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
		case *request2.MediaRequest:
			id = strconv.Itoa(v.NmId)
		}
		if _, isBanned := bannedSet[id]; !isBanned {
			filteredModels = append(filteredModels, model)
		}
	}

	return filteredModels
}
