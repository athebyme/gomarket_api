package handlers

import (
	"encoding/json"
	"gomarketplace_api/internal/wholesaler/internal/models/requests"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"gomarketplace_api/pkg/dbconnect"
	"io"
	"log"
	"net/http"
	"time"
)

type BrandHandler struct {
	dbconnect.Database
	repo *repositories.BrandRepository
}

func NewBrandHandler(connector dbconnect.Database, writer io.Writer) *BrandHandler {
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

	productRepo := repositories.NewProductRepository(db, productUpdater)
	return &BrandHandler{
		Database: connector,
		repo:     repositories.NewBrandRepository(productRepo),
	}
}

func (h *BrandHandler) GetBrandHandler(w http.ResponseWriter, r *http.Request) {

	if err := h.Ping(); err != nil {
		http.Error(w, "Failed to ping database", http.StatusInternalServerError)
		return
	}

	// Декодирование тела запроса
	var req requests.BrandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var brands map[int]string
	var err error

	startTime := time.Now()
	if len(req.ProductIDs) == 0 {
		brands, err = h.repo.GetProductsBrands()
		if err != nil {
			http.Error(w, "Failed to fetch all brands", http.StatusInternalServerError)
			return
		}
	} else {
		brands, err = h.repo.GetProductBrandByIDs(req.ProductIDs)
		if err != nil {
			http.Error(w, "Failed to fetch brands", http.StatusInternalServerError)
			return
		}
	}
	log.Printf("brands handler response execution time: %v", time.Since(startTime))

	err = json.NewEncoder(w).Encode(brands)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
