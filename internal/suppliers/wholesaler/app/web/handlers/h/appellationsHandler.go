package h

import (
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/business"
	requests2 "gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	"log"
	"net/http"
)

type AppellationHandler struct {
	productService *business.ProductService
}

func NewAppellationHandler(productService *business.ProductService) *AppellationHandler {
	return &AppellationHandler{
		productService: productService,
	}
}

func (h *AppellationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	var appellReq requests2.AppellationsRequest
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
