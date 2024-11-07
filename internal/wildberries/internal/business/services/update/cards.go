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
	"log"
	"net/http"
	"sync"
	"time"
)

type CardUpdater struct {
	NomenclatureService get.NomenclatureService
	wsclient            *clients2.WServiceClient
	textService         service.ITextService
}

func NewCardUpdater(nservice *get.NomenclatureService, textService service.ITextService, wsClientUrl string) *CardUpdater {
	return &CardUpdater{
		NomenclatureService: *nservice,
		wsclient:            clients2.NewWServiceClient(wsClientUrl),
		textService:         textService,
	}
}

const updateCardsUrl = "https://content-api.wildberries.ru/content/v2/cards/update"

func UpdateCards() (int, error) {
	panic("TO DO")
}
func (u *CardUpdater) UpdateCardNaming(settings request.Settings) (int, error) {
	const UPLOAD_SIZE = 300
	const MaxBatchSize = 1 << 20 // 1 MB
	const GOROUTINE_COUNT = 5
	const REQUEST_RATE_LIMIT = 50 // 100 запросов в минуту = max
	const UPLOAD_RATE_LIMIT = 5
	var updatedCount = 0
	var currentBatch []models.WildberriesCard
	var currentBatchSize int
	var gotData []response.Nomenclature

	log.Println("Fetching appellations...")
	appellationsMap, err := u.wsclient.FetchAppellations()
	if err != nil {
		return 0, fmt.Errorf("error fetching appellations: %w", err)
	}

	log.Println("Fetching descriptions...")
	descriptionsMap, err := u.wsclient.FetchDescriptions()
	if err != nil {
		return 0, fmt.Errorf("error fetching descriptions: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var processedItems sync.Map
	nomenclatureChan := make(chan response.Nomenclature)
	uploadChan := make(chan []models.WildberriesCard) // Канал для отправки данных

	responseLimiter := rate.NewLimiter(rate.Every(time.Minute/REQUEST_RATE_LIMIT), REQUEST_RATE_LIMIT)
	uploadToServerLimiter := rate.NewLimiter(rate.Every(time.Minute/UPLOAD_RATE_LIMIT), REQUEST_RATE_LIMIT)

	// Запуск горутин для обработки номенклатур
	log.Println("Starting goroutines for processing nomenclatures...")
	for i := 0; i < GOROUTINE_COUNT; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for nomenclature := range nomenclatureChan {
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

				wbCard.Title = u.textService.ClearAndReduce(appellationsMap[globalId], 60)
				wbCard.Title = u.textService.ReplaceEngLettersToRus(wbCard.Title)
				if description, ok := descriptionsMap[globalId]; ok && description != "" {
					wbCard.Description = u.textService.ClearAndReduce(description, 2000)
				} else {
					wbCard.Description = u.textService.ClearAndReduce(appellationsMap[globalId], 2000)
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
	var uploadWg sync.WaitGroup
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

			cards, err := uploadCards(batch)
			if err != nil {
				log.Printf("Error during uploading %s", err)
				return
			}
			updatedCount += cards
		}
	}()

	// Получение номенклатур конкурентно
	log.Println("Fetching and sending nomenclatures to the channel...")
	go func() {
		if err := u.NomenclatureService.GetNomenclaturesWithLimitConcurrentlyPutIntoChanel(settings.Cursor.Limit, "", nomenclatureChan, responseLimiter); err != nil {
			log.Printf("Error fetching nomenclatures concurrently: %s", err)
		}
		close(nomenclatureChan)
		wg.Wait()

		// Обработка оставшихся данных
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

	uploadWg.Wait()

	log.Printf("Update completed, total updated count: %d", updatedCount)
	return updatedCount, nil

}

func uploadCards(cards []models.WildberriesCard) (int, error) {
	log.Printf("Sending updated card to Wildberries...")
	requestBody, err := json.Marshal(cards)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal update request: %w", err)
	}

	var jsonCheck []map[string]interface{}
	if err := json.Unmarshal(requestBody, &jsonCheck); err != nil {
		log.Printf("Invalid JSON: %v", err)
		return 0, fmt.Errorf("invalid JSON: %w", err)
	}

	req, err := http.NewRequest("POST", updateCardsUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create update request: %w", err)
	}

	requestBodySize := len(requestBody) // Перевод размера в мегабайты
	requestBodySizeMB := float64(requestBodySize) / (1 << 20)
	log.Printf("Request Body Size: %d bytes (%.2f MB)\n", requestBodySize, requestBodySizeMB)

	req.Header.Set("Content-Type", "application/json")
	services.SetAuthorizationHeader(req)

	log.Printf("Sending requestbody: %v", string(requestBody))
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to update cards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	// Обновление истории карт

	return len(cards), nil
}
