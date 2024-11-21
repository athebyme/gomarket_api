package handlers

import (
	"encoding/json"
	"gomarketplace_api/internal/wholesaler/internal/business"
	"gomarketplace_api/internal/wholesaler/internal/models/requests"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/pkg/dbconnect"
	"log"
	"net/http"
	"time"
)

type PriceHandler struct {
	dbconnect.DbConnector
	business.PriceService
}

func NewPriceHandler(connector dbconnect.DbConnector) *PriceHandler {
	db, err := connector.Connect()
	if err != nil {
		return nil
	}
	priceSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_prices.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_prices.csv",
		LastUpdateColumn: "last_update_prices"}

	priceUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"price",
		[]string{"id товара", "цена"},
		priceSource.LastUpdateColumn,
		priceSource.InfURL,
		priceSource.CSVURL)

	priceRepo := storage.NewPriceRepository(db, priceUpdater)
	err = priceRepo.Update([]string{"global_id", "price"})
	if err != nil {
		return nil
	}

	priceService := business.NewPriceEngine(priceRepo)

	return &PriceHandler{
		DbConnector:  connector,
		PriceService: priceService,
	}
}

func (h *PriceHandler) GetPriceHandler(w http.ResponseWriter, r *http.Request) {

	if err := h.Ping(); err != nil {
		http.Error(w, "Failed to ping database", http.StatusInternalServerError)
		return
	}

	// Декодирование тела запроса
	var priceRequest requests.PriceRequest
	if err := json.NewDecoder(r.Body).Decode(&priceRequest); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var response interface{}
	var err error

	startTime := time.Now()
	if len(priceRequest.ProductIDs) == 0 { // Проверка наличия productIDs
		response, err = h.GetPrices(priceRequest.All)
		if err != nil {
			http.Error(w, "Failed to fetch all media sources", http.StatusInternalServerError)
			return
		}
	} else {
		response, err = h.GetPriceById(priceRequest.ProductIDs[0])
		if err != nil {
			http.Error(w, "Failed to fetch media sources", http.StatusInternalServerError)
			return
		}
	}
	log.Printf("prices source execution time: %v", time.Since(startTime))

	// Кодирование ответа
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
