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
	var updatedCount = 0
	var currentBatch []models.WildberriesCard
	var currentBatchSize int

	// Список всех global ids в wholesaler.products
	appellationsMap, err := u.wsclient.FetchAppellations()
	if err != nil {
		return 0, fmt.Errorf("error fetching appellations: %w", err)
	}
	descriptionsMap, err := u.wsclient.FetchDescriptions()
	if err != nil {
		return 0, fmt.Errorf("error fetching descriptions: %w", err)
	}

	r, b := 3, 3

	limiter := rate.NewLimiter(rate.Limit(r), b)

	var wg sync.WaitGroup
	var mu sync.Mutex
	nomenclatureChan := make(chan response.Nomenclature)

	// Запуск горутин для обработки номенклатур
	for i := 0; i < r; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for nomenclature := range nomenclatureChan {
				var wbCard models.WildberriesCard
				wbCard = *wbCard.FromNomenclature(nomenclature)

				globalId, err := nomenclature.GlobalID()
				if err != nil || globalId == 0 {
					log.Printf("Updating naming | (globalID=%s) parse error (not SPB aricular)", nomenclature.VendorCode)
					continue
				}
				if _, ok := appellationsMap[globalId]; !ok {
					log.Printf("Updating naming | (globalID=%s) not found appellation", nomenclature.VendorCode)
					continue
				}

				// wbCard.Title = u.textService.AddWordIfNotExistsToFront(appellationsMap[globalId], nomenclature.SubjectName)
				wbCard.Title = u.textService.ClearAndReduce(appellationsMap[globalId], 60)
				//wbCard.Title = u.textService.RemoveWord(wbCard.Title, wbCard.Brand) // УДАЛЯЕМ БРЕНД ИЗ НАЗВАНИЯ ( УВЕЛИЧИВАЕТ КАЧЕСТВО КАРТОЧКИ ! )
				wbCard.Title = u.textService.ReplaceEngLettersToRus(wbCard.Title)
				if description, ok := descriptionsMap[globalId]; ok && description != "" {
					wbCard.Description = u.textService.ClearAndReduce(description, 2000)
				} else {
					wbCard.Description = u.textService.ClearAndReduce(appellationsMap[globalId], 2000)
				}

				mu.Lock()
				currentBatch = append(currentBatch, wbCard)
				currentBatchSize += len([]byte(wbCard.Title)) + len([]byte(wbCard.Description))
				if len(currentBatch) >= UPLOAD_SIZE || currentBatchSize >= 1<<20 {
					cards, err := uploadCards(currentBatch)
					if err != nil {
						log.Printf("Error during uploading %s", err)
						// err
						return
					}
					updatedCount += cards

					currentBatch = nil
					currentBatchSize = 0
				}
				mu.Unlock()
			}
		}()
	}

	// Получение номенклатур и отправка их в канал для обработки пакетами по 100
	packetSizes := divideLimitsToPackets(settings.Cursor.Limit, 100)
	for _, limit := range packetSizes {
		settings.Cursor.Limit = limit
		if err := limiter.Wait(context.Background()); err != nil {
			return 0, err
		}

		nomenclatureResponse, err := u.NomenclatureService.GetNomenclatures(settings, "")
		if err != nil {
			close(nomenclatureChan)
			return 0, fmt.Errorf("failed to get nomenclatures: %w", err)
		}

		if len(nomenclatureResponse.Data) == 0 {
			break
		} else {
			settings.Cursor.UpdatedAt = nomenclatureResponse.Data[len(nomenclatureResponse.Data)-1].UpdatedAt
			settings.Cursor.NmID = nomenclatureResponse.Data[len(nomenclatureResponse.Data)-1].NmID
		}

		for _, nomenclature := range nomenclatureResponse.Data {
			nomenclatureChan <- nomenclature
		}
	}

	close(nomenclatureChan)
	wg.Wait()

	// Добавляем оставшиеся данные в cardsToUpdate
	if len(currentBatch) > 0 {
		cards, err := uploadCards(currentBatch)
		if err != nil {
			return 0, err
		}
		updatedCount += cards
	}

	return updatedCount, nil
}

func uploadCards(cards []models.WildberriesCard) (int, error) {
	log.Printf("Sending updated card to Wildberries...")
	requestBody, err := json.Marshal(cards)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal update request: %w", err)
	}

	req, err := http.NewRequest("POST", updateCardsUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	services.SetAuthorizationHeader(req)

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

func divideLimitsToPackets(totalCount int, packetSize int) []int {
	var packets []int
	for count := totalCount; count > 0; count -= packetSize {
		if count >= packetSize {
			packets = append(packets, packetSize)
		} else {
			packets = append(packets, count)
		}
	}
	return packets
}
