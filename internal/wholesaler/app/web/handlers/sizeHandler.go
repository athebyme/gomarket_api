package handlers

import (
	"database/sql"
	"encoding/json"
	"gomarketplace_api/internal/wholesaler/internal/business"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"gomarketplace_api/pkg/dbconnect"
	"gomarketplace_api/pkg/logger"
	"io"
	"net/http"
	"time"
)

type SizeHandler struct {
	dbconnect.Database
	service *business.SizeService
	logger  logger.Logger
}

func NewSizeHandler(connector dbconnect.Database, writer io.Writer) *SizeHandler {
	_log := logger.NewLogger(writer, "[SizeHandler]")

	db, err := connector.Connect()
	if err != nil {
		return nil
	}

	//TO DO
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
	//

	sizeRepo := repositories.NewSizeRepository(db, productUpdater, writer)
	sizeService := business.NewSizeService(sizeRepo, writer)
	return &SizeHandler{
		service:  sizeService,
		Database: connector,
		logger:   _log,
	}
}

func (h *SizeHandler) GetSizeHandler(w http.ResponseWriter, r *http.Request) {
	var sizes map[int][]models.SizeWrapper
	var err error

	startTime := time.Now()
	sizes, err = h.service.GetSizes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	h.logger.Log("media source execution time: %v", time.Since(startTime))

	// Кодирование ответа
	err = json.NewEncoder(w).Encode(sizes)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
func (h *SizeHandler) Connect() (*sql.DB, error) {
	return h.Database.Connect()
}

func (h *SizeHandler) Ping() error {
	err := h.Database.Ping()
	if err != nil {
		return err
	}
	return nil
}
