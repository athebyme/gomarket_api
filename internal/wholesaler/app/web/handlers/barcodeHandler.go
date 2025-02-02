package handlers

import (
	"encoding/json"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"gomarketplace_api/internal/wholesaler/pkg/requests"
	"gomarketplace_api/pkg/dbconnect"
	"gomarketplace_api/pkg/logger"
	"io"
	"net/http"
	"time"
)

type BarcodeHandler struct {
	dbconnect.Database
	repo *repositories.BarcodeRepository
	log  logger.Logger
}

func NewBarcodeHandler(con dbconnect.Database, writer io.Writer) *BarcodeHandler {
	_log := logger.NewLogger(writer, "[BarcodeHandler]")

	db, err := con.Connect()
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
			"sex", "color", "dimension", "package", "empty",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := repositories.NewProductRepository(db, productUpdater)
	barcodeRepo := repositories.NewBarcodeRepository(productRepo)

	return &BarcodeHandler{
		Database: con,
		repo:     barcodeRepo,
		log:      _log,
	}
}

func (h *BarcodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if err := h.Ping(); err != nil {
		http.Error(w, "Failed to ping database", http.StatusInternalServerError)
		return
	}

	// Декодирование тела запроса
	var barcodeReq requests.BarcodeRequest
	if err := json.NewDecoder(r.Body).Decode(&barcodeReq); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var barcodesResponse map[int]interface{}
	var err error

	startTime := time.Now()
	if len(barcodeReq.ProductIDs) == 0 { // Проверка наличия productIDs
		barcodesResponse, err = h.repo.GetAllBarcodes()
		if err != nil {
			http.Error(w, "Failed to fetch all media sources", http.StatusInternalServerError)
			return
		}
	} else {
		barcodesResponse, err = h.repo.GetBarcodesByProductIDs(barcodeReq.ProductIDs)
		if err != nil {
			http.Error(w, "Failed to fetch media sources", http.StatusInternalServerError)
			return
		}
	}
	h.log.Log("media handler response execution time: %v", time.Since(startTime))

	// Кодирование ответа
	err = json.NewEncoder(w).Encode(barcodesResponse)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
