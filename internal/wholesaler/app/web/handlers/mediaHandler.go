package handlers

import (
	"encoding/json"
	"gomarketplace_api/internal/wholesaler/internal/models/requests"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/pkg/dbconnect"
	"log"
	"net/http"
	"time"
)

type MediaHandler struct {
	dbconnect.DbConnector
	repo *storage.MediaRepository
}

func NewMediaHandler(connector dbconnect.DbConnector) *MediaHandler {
	db, err := connector.Connect()
	if err != nil {
		return nil
	}
	productSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_info.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_info.csv",
		LastUpdateColumn: "last_update_products"}

	productUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"product",
		[]string{"global_id", "model", "appellation", "category",
			"brand", "country", "product_type", "features",
			"sex", "color", "dimension", "package",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := storage.NewProductRepository(db, productUpdater)
	return &MediaHandler{
		DbConnector: connector,
		repo:        storage.NewMediaRepository(productRepo),
	}
}

func (h *MediaHandler) GetMediaHandler(w http.ResponseWriter, r *http.Request) {

	if err := h.Ping(); err != nil {
		http.Error(w, "Failed to ping database", http.StatusInternalServerError)
		return
	}

	// Декодирование тела запроса
	var mediaReq requests.MediaRequest
	if err := json.NewDecoder(r.Body).Decode(&mediaReq); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var mediaMap map[int][]string
	var err error

	startTime := time.Now()
	if len(mediaReq.ProductIDs) == 0 { // Проверка наличия productIDs
		mediaMap, err = h.repo.GetMediaSources(mediaReq.Censored)
		if err != nil {
			http.Error(w, "Failed to fetch all media sources", http.StatusInternalServerError)
			return
		}
	} else {
		mediaMap, err = h.repo.GetMediaSourcesByProductIDs(mediaReq.ProductIDs, mediaReq.Censored)
		if err != nil {
			http.Error(w, "Failed to fetch media sources", http.StatusInternalServerError)
			return
		}
	}
	log.Printf("media source execution time: %v", time.Since(startTime))

	// Кодирование ответа
	err = json.NewEncoder(w).Encode(mediaMap)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}