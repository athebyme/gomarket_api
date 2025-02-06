package h

import (
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/business"
	requests2 "gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	"log"
	"net/http"
)

type DescriptionsHandler struct {
	productService *business.ProductService
}

func NewDescriptionsHandler(productService *business.ProductService) *DescriptionsHandler {
	return &DescriptionsHandler{
		productService: productService,
	}
}

func (h *DescriptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	var descReq requests2.DescriptionRequest
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
