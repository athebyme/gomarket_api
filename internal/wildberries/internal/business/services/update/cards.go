package update

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	get2 "gomarketplace_api/internal/wildberries/internal/business/services/get"
	clients2 "gomarketplace_api/internal/wildberries/pkg/clients"
	"gomarketplace_api/pkg/business/service"
	"log"
	"net/http"
	"time"
)

type CardUpdater struct {
	NomenclatureService get2.NomenclatureUpdateGetter
	wsclient            *clients2.WServiceClient
	textService         service.ITextService
}

func NewCardUpdater(nservice *get2.NomenclatureUpdateGetter, textService service.ITextService, wsClientUrl string) *CardUpdater {
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
	var cardsToUpdate []get.WildberriesCard

	// список всех global ids в wholesaler.products
	appellationsMap, err := u.wsclient.FetchAppellations()
	descriptionsMap, err := u.wsclient.FetchDescriptions()
	if err != nil {
		log.Fatalf("Error fetching Global IDs: %s", err)
	}

	var r, b int
	if settings.Cursor.Limit < 20 {
		r, b = 5, 5
	} else if settings.Cursor.Limit < 50 {
		r, b = 2, 2
	} else {
		r, b = 1, 1
	}

	updated := 0

	limiter := rate.NewLimiter(rate.Limit(r), b)
	if err := limiter.Wait(context.Background()); err != nil {
		return updated, err
	}
	nomenclatureResponse, err := u.NomenclatureService.GetNomenclature(settings, "")
	if err != nil {
		return updated, fmt.Errorf("failed to get nomenclatures: %w", err)
	}

	for _, v := range nomenclatureResponse.Data {
		var wbCard get.WildberriesCard
		wbCard = *wbCard.FromNomenclature(v)

		globalId, err := v.GlobalID()
		if err != nil {
			log.Printf("(globalID=%s) parse error (not SPB aricular)", v.VendorCode)
			continue
		}
		if globalId == 0 {
			log.Printf("(globalID=%s) parse error (not SPB aricular)", v.VendorCode)
			continue
		}
		if _, ok := appellationsMap[globalId]; !ok {
			log.Printf("(globalID=%s) not found appellation", v.VendorCode)
			continue
		}

		// wbCard.Title = u.textService.AddWordIfNotExistsToFront(appellationsMap[globalId], v.SubjectName)
		wbCard.Title = u.textService.ClearAndReduce(appellationsMap[globalId], 60)
		//wbCard.Title = u.textService.RemoveWord(wbCard.Title, wbCard.Brand) // УДАЛЯЕМ БРЕНД ИЗ НАЗВАНИЯ ( УВЕЛИЧИВАЕТ КАЧЕСТВО КАРТОЧКИ ! )
		wbCard.Title = u.textService.ReplaceEngLettersToRus(wbCard.Title)
		if description, ok := descriptionsMap[globalId]; ok && description != "" {
			wbCard.Description = u.textService.ClearAndReduce(description, 2000)
		} else {
			wbCard.Description = u.textService.ClearAndReduce(appellationsMap[globalId], 2000)
		}
		cardsToUpdate = append(cardsToUpdate, wbCard)
	}

	requestBody, err := json.Marshal(cardsToUpdate)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal update request: %w", err)
	}
	log.Printf("Updating cards: %s", string(requestBody))

	req, err := http.NewRequest("POST", updateCardsUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	services.SetAuthorizationHeader(req)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to update cards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	// update card history

	return len(cardsToUpdate), nil
}
