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
	"gomarketplace_api/internal/wildberries/pkg/clients"
	"log"
	"net/http"
	"time"
)

const updateCardsUrl = "https://content-api.wildberries.ru/content/v2/cards/update"

func UpdateCards() (int, error) {
	panic("TO DO")
}

// map = vendorCode -> appellation
func UpdateCardAppellation(settings request.Settings) (int, error) {
	var cardsToUpdate []get.WildberriesCard
	clientWS := clients.NewAppellationsClient("http://localhost:8081")

	// список всех global ids в wholesaler.products
	appellationsMap, err := clientWS.FetchAppellations()
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
	nomenclatureResponse, err := get2.GetNomenclature(settings, "")
	if err != nil {
		return updated, fmt.Errorf("failed to get nomenclatures: %w", err)
	}

	for _, v := range nomenclatureResponse.Data {
		var wbCard get.WildberriesCard
		wbCard = *wbCard.FromNomenclature(v)

		globalId, err := v.GlobalID()
		if err != nil {
			continue
		}
		if globalId == 0 {
			continue
		}
		if _, ok := appellationsMap[globalId]; !ok {
			continue
		}

		wbCard.Title = appellationsMap[globalId]
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

	return len(cardsToUpdate), nil
}
