package update

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	models "gomarketplace_api/internal/wildberries/internal/business/models/get"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	clients2 "gomarketplace_api/internal/wildberries/pkg/clients"
	"gomarketplace_api/pkg/business/service"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type CardUpdater struct {
	NomenclatureService get.NomenclatureEngine
	wsclient            *clients2.WServiceClient
	textService         service.ITextService
	services.AuthEngine
}

func NewCardUpdater(nservice *get.NomenclatureEngine, textService service.ITextService, wsClientUrl string, auth services.AuthEngine) *CardUpdater {
	return &CardUpdater{
		NomenclatureService: *nservice,
		wsclient:            clients2.NewWServiceClient(wsClientUrl),
		textService:         textService,
		AuthEngine:          auth,
	}
}

const updateCardsUrl = "https://content-api.wildberries.ru/content/v2/cards/update"

func UpdateCards() (int, error) {
	panic("TO DO")
}
func (cu *CardUpdater) UpdateCardNaming(settings request.Settings) (int, error) {
	const UPLOAD_SIZE = 100
	const MaxBatchSize = 1 << 20 // 1 MB
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 50 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 10
	var updatedCount = 0
	var currentBatch []models.WildberriesCard
	var currentBatchSize int
	var gotData []response.Nomenclature
	var goroutinesNmsCount atomic.Int32

	log.Println("Fetching appellations...")
	appellationsMap, err := cu.wsclient.FetchAppellations()
	if err != nil {
		return 0, fmt.Errorf("error fetching appellations: %w", err)
	}

	log.Println("Fetching descriptions...")
	descriptionsMap, err := cu.wsclient.FetchDescriptions()
	if err != nil {
		return 0, fmt.Errorf("error fetching descriptions: %w", err)
	}

	var processWg sync.WaitGroup
	var mu sync.Mutex
	var processedItems sync.Map
	nomenclatureChan := make(chan response.Nomenclature)
	uploadChan := make(chan []models.WildberriesCard) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), UPLOAD_RATE_LIMIT)

	log.Println("Fetching and sending nomenclatures to the channel...")
	go func() {
		if err := cu.NomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChanel(settings, "", nomenclatureChan, responseLimiter); err != nil {
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
					continue
				}
				if _, ok := appellationsMap[globalId]; !ok {
					log.Printf("(G%d) (globalID=%s) not found appellation", i, nomenclature.VendorCode)
					continue
				}

				wbCard.Title = cu.textService.ClearAndReduce(appellationsMap[globalId], 60)
				changedBrand := cu.textService.ReplaceEngLettersToRus(" " + wbCard.Brand)
				wbCard.Title = cu.textService.FitIfPossible(wbCard.Title, changedBrand, 60)

				if description, ok := descriptionsMap[globalId]; ok && description != "" {
					wbCard.Description = cu.textService.ClearAndReduce(description, 2000)
				} else {
					wbCard.Description = cu.textService.ClearAndReduce(appellationsMap[globalId], 2000)
				}

				mu.Lock()
				gotData = append(gotData, nomenclature)
				currentBatch = append(currentBatch, wbCard)
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
	go func() {
		for batch := range uploadChan {
			log.Println("Uploading batch of cards...")
			// Лимитирование запросов на загрузку
			if err := uploadToServerLimiter.Wait(context.Background()); err != nil {
				log.Printf("Error waiting for rate limiter: %s", err)
				return
			}

			cards, err := cu.processAndUploadCards(batch)
			if err != nil {
				log.Printf("Error during uploading %s", err)
				return
			}
			updatedCount += cards
		}
	}()

	go func() {
		processWg.Wait()

		// Process remaining data
		log.Println("Processing any remaining data...")
		mu.Lock()
		if len(currentBatch) > 0 {
			log.Println("Uploading remaining batch of cards...")
			uploadChan <- currentBatch
			currentBatch = nil
		}
		mu.Unlock()
		close(uploadChan)
	}()

	log.Printf("Goroutines fetchers got (%d) nomenclautres", goroutinesNmsCount.Load())
	log.Printf("Update completed, total updated count: %d", updatedCount)
	return updatedCount, nil

}

func (cu *CardUpdater) processAndUploadCards(cards []models.WildberriesCard) (int, error) {
	bodyBytes, statusCode, err := cu.uploadCards(cards)
	if err != nil {
		if statusCode != http.StatusOK && bodyBytes != nil {
			var errorResponse map[string]interface{}
			log.Printf("Trying to fix (Status=%d)...", statusCode)
			if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil {
				if additionalErrors, ok := errorResponse["additionalErrors"].(map[string]interface{}); ok {
					if bannedArticles, ok := additionalErrors["Забаненные артикулы WB"].(string); ok {
						bannedArticlesSlice := strings.Split(bannedArticles, ", ")
						filteredCards := filterOutBannedArticles(cards, bannedArticlesSlice)
						if len(filteredCards) > 0 {
							return cu.processAndUploadCards(filteredCards)
						}
					}
				}
			}
		}
		return 0, fmt.Errorf("update failed: %w", err)
	}

	return len(cards), nil
}

func (cu *CardUpdater) uploadCards(cards []models.WildberriesCard) ([]byte, int, error) {
	log.Printf("Sending updated card to Wildberries...")
	requestBody, err := json.Marshal(cards)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to marshal update request: %w", err)
	}

	var jsonCheck []map[string]interface{}
	if err := json.Unmarshal(requestBody, &jsonCheck); err != nil {
		log.Printf("Invalid JSON: %v", err)
		return nil, 500, fmt.Errorf("invalid JSON: %w", err)
	}

	req, err := http.NewRequest("POST", updateCardsUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, 500, fmt.Errorf("failed to create update request: %w", err)
	}

	requestBodySize := len(requestBody)
	requestBodySizeMB := float64(requestBodySize) / (1 << 20)
	log.Printf("Request Body Size: %d bytes (%.2f MB)\n", requestBodySize, requestBodySizeMB)

	req.Header.Set("Content-Type", "application/json")
	cu.SetApiKey(req)

	log.Printf("Sending requestbody: %v", string(requestBody))
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to update cards: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read error response: %w", err)
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

		log.Printf("Update failed with status: %d, error details: %s", resp.StatusCode, string(errorDetails))
		return bodyBytes, resp.StatusCode, fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	// Обновление истории карт
	return bodyBytes, resp.StatusCode, nil
}

func filterOutBannedArticles(cards []models.WildberriesCard, bannedArticles []string) []models.WildberriesCard {
	var filteredCards []models.WildberriesCard
	bannedSet := make(map[string]struct{}, len(bannedArticles))
	for _, article := range bannedArticles {
		bannedSet[article] = struct{}{}
	}

	for _, card := range cards {
		if _, isBanned := bannedSet[strconv.Itoa(card.NmID)]; !isBanned {
			filteredCards = append(filteredCards, card)
		}
	}
	return filteredCards
}
