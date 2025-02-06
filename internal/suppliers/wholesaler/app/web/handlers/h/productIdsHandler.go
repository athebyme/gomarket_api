package h

import (
	"database/sql"
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/business"
	"gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"
	"log"
	"net/http"
)

type WholesalerIdsHandler struct {
	productService *business.ProductService
}

func NewWholesalerIdsHandler(db *sql.DB) *WholesalerIdsHandler {
	productRepo := repositories.NewProductRepository(db)
	productService := business.NewProductService(productRepo)

	return &WholesalerIdsHandler{
		productService: productService,
	}
}

func (h *WholesalerIdsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
