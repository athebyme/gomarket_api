package handlers

import (
	"database/sql"
	"encoding/json"
	"gomarketplace_api/internal/wholesaler/internal/business"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"gomarketplace_api/internal/wholesaler/pkg/requests"
	"gomarketplace_api/pkg/dbconnect"
	"log"
	"net/http"
)

type ProductHandler struct {
	dbconnect.Database
	productService *business.ProductService
}

func NewProductHandler(connector dbconnect.Database) *ProductHandler {
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
			"sex", "color", "dimension", "package", "empty",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := repositories.NewProductRepository(db, productUpdater)

	err = productRepo.Update()
	if err != nil {
		return nil
	}

	productService := business.NewProductService(productRepo)

	return &ProductHandler{
		Database:       connector,
		productService: productService,
	}
}

func (h *ProductHandler) Connect() (*sql.DB, error) {
	return h.Database.Connect()
}

func (h *ProductHandler) Ping() error {
	err := h.Database.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (h *ProductHandler) GetGlobalIDsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := h.Connect()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	if err = db.Ping(); err != nil {
		return
	}

	globalIDs, err := h.productService.GetAllGlobalIDs()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(globalIDs)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func (h *ProductHandler) GetAppellationHandler(w http.ResponseWriter, r *http.Request) {
	db, err := h.Connect()
	if err != nil {
		return
	}

	if err = db.Ping(); err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	var appellReq requests.AppellationsRequest
	if err = json.NewDecoder(r.Body).Decode(&appellReq); err != nil {
		http.Error(w, "Failed to fetch Appellations", http.StatusInternalServerError)
	}

	var appellations map[int]interface{}
	if len(appellReq.ProductIDs) == 0 {
		appellations, err = h.productService.GetAllAppellations()
	} else {
		appellations, err = h.productService.GetAllAppellationsByIDs(appellReq.ProductIDs)

	}
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(appellations)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func (h *ProductHandler) GetDescriptionsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := h.Connect()
	if err != nil {
		return
	}

	if err = db.Ping(); err != nil {
		return
	}

	var descReq requests.DescriptionRequest
	if err = json.NewDecoder(r.Body).Decode(&descReq); err != nil {
		http.Error(w, "Failed to fetch Descriptions", http.StatusInternalServerError)
	}

	var descriptions map[int]interface{}

	if len(descReq.ProductIDs) == 0 {
		descriptions, err = h.productService.GetAllDescriptions(descReq.IncludeEmptyDescriptions)
	} else {
		descriptions, err = h.productService.GetAllDescriptionsByIDs(descReq.ProductIDs, descReq.IncludeEmptyDescriptions)
	}

	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(descriptions)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}
