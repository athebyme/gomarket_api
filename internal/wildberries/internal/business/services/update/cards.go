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
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	models "gomarketplace_api/internal/wildberries/internal/business/models/get"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	"gomarketplace_api/internal/wildberries/internal/business/services/parse"
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

type CardUpdateService struct {
	nomenclatureService get.SearchEngine
	wsclient            *clients2.WServiceClient
	textService         service.ITextService
	brandService        parse.BrandService
	defaultValues       values.WildberriesValues
	services.AuthEngine
}

func NewCardUpdateService(nservice *get.SearchEngine, textService service.ITextService, wsClientUrl string, auth services.AuthEngine, writer io.Writer, brandService parse.BrandService, wbDefaultValues values.WildberriesValues) *CardUpdateService {
	return &CardUpdateService{
		nomenclatureService: *nservice,
		wsclient:            clients2.NewWServiceClient(wsClientUrl, writer),
		textService:         textService,
		brandService:        brandService,
		AuthEngine:          auth,
		defaultValues:       wbDefaultValues,
	}
}

const updateCardsUrl = "https://content-api.wildberries.ru/content/v2/cards/update"

var numberOfErroredNomenclatures atomic.Int32
var updatedCount atomic.Int32

func UpdateCards() (int, error) {
	panic("TO DO")
}

func (cu *CardUpdateService) UpdateCardNaming(settings request.Settings) (int, error) {
	const UPLOAD_SIZE = 2000
	const MaxBatchSize = 1 << 20 // 1 MB
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 70 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 10
	var currentBatch []request.Model
	var currentBatchSize int
	var gotData []response.Nomenclature
	var goroutinesNmsCount atomic.Int32

	log.Println("Fetching filterAppellations...")
	appellationsMap, err := cu.wsclient.FetchAppellations(requests.AppellationsRequest{FilterRequest: requests.FilterRequest{ProductIDs: []int{}}})
	if err != nil {
		return 0, fmt.Errorf("error fetching filterAppellations: %w", err)
	}

	log.Println("Fetching descriptions...")
	descriptionsMap, err := cu.wsclient.FetchDescriptions(requests.AppellationsRequest{FilterRequest: requests.FilterRequest{ProductIDs: []int{}}})
	if err != nil {
		return 0, fmt.Errorf("error fetching descriptions: %w", err)
	}

	var processWg sync.WaitGroup
	var uploadWg sync.WaitGroup
	var mu sync.Mutex
	var processedItems sync.Map
	nomenclatureChan := make(chan response.Nomenclature)
	uploadChan := make(chan []request.Model) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)

	log.Println("Fetching and sending nomenclatures to the channel...")
	go func() {
		if err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChanel(settings, "", nomenclatureChan, responseLimiter); err != nil {
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
				if _, ok := appellationsMap[globalId]; !ok {
					log.Printf("(G%d) (globalID=%s) not found appellation", i, nomenclature.VendorCode)
					numberOfErroredNomenclatures.Add(1)
					continue
				}

				wbCard.Title = cu.textService.ClearAndReduce(appellationsMap[globalId].(string), 60)
				changedBrand := cu.textService.ReplaceEngLettersToRus(" " + wbCard.Brand)
				wbCard.Title = cu.textService.FitIfPossible(wbCard.Title, changedBrand, 60)

				if description, ok := descriptionsMap[globalId]; ok && description != "" {
					wbCard.Description = cu.textService.ClearAndReduce(description.(string), 2000)
				} else {
					wbCard.Description = cu.textService.ClearAndReduce(appellationsMap[globalId].(string), 2000)
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

func (cu *CardUpdateService) UpdateCardPackages(settings request.Settings) (int, error) {
	const UPLOAD_SIZE = 2000
	const MaxBatchSize = 1 << 20 // 1 MB
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 70 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 10
	var currentBatch []request.Model
	var currentBatchSize int
	var gotData []response.Nomenclature
	var goroutinesNmsCount atomic.Int32

	var processWg sync.WaitGroup
	var uploadWg sync.WaitGroup
	var mu sync.Mutex
	var processedItems sync.Map
	nomenclatureChan := make(chan response.Nomenclature)
	uploadChan := make(chan []request.Model) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)

	log.Println("Fetching and sending nomenclatures to the channel...")
	go func() {
		if err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChanel(settings, "", nomenclatureChan, responseLimiter); err != nil {
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

				wbCard.Dimensions = response.DimensionWrapper{
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

func (cu *CardUpdateService) UpdateCardMedia(settings request.Settings) (int, error) {
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

	nomenclatureChan := make(chan response.Nomenclature)
	uploadChan := make(chan request.Model) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)

	log.Println("Fetching and sending nomenclatures to the channel...")
	go func() {
		if err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChanel(settings, "", nomenclatureChan, responseLimiter); err != nil {
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

				var urls []string
				var ok bool
				if urls, ok = mediaMap[globalId]; !ok && len(urls) < 0 {
					log.Printf("(G%d) (globalID=%s) parse error (not SPB aricular)", i, nomenclature.VendorCode)
					numberOfErroredNomenclatures.Add(1)
					continue
				}

				mediaRequest := request.NewMediaRequest(nomenclature.NmID, urls)

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

func (cu *CardUpdateService) UpdateCardBrand(settings request.Settings) (int, error) {
	const UPLOAD_SIZE = 2000
	const MaxBatchSize = 1 << 20 // 1 MB
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 70 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 10
	var currentBatch []request.Model
	var currentBatchSize int
	var gotData []response.Nomenclature
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
	nomenclatureChan := make(chan response.Nomenclature)
	uploadChan := make(chan []request.Model) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)

	log.Println("Fetching and sending nomenclatures to the channel...")
	go func() {
		if err := cu.nomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChanel(settings, "", nomenclatureChan, responseLimiter); err != nil {
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

func (cu *CardUpdateService) UpdateDBNomenclatures(settings request.Settings, locale string) (int, error) {
	return cu.nomenclatureService.UploadToDb(settings, locale)
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
		case *request.MediaRequest:
			id = strconv.Itoa(v.NmId)
		}
		if _, isBanned := bannedSet[id]; !isBanned {
			filteredModels = append(filteredModels, model)
		}
	}

	return filteredModels
}
