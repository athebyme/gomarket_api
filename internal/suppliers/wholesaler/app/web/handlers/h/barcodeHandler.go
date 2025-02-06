package h

import (
	"database/sql"
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	"gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"
	"net/http"
)

type BarcodeHandler struct {
	repo *repositories.BarcodeRepository
}

func NewBarcodeHandler(db *sql.DB) *BarcodeHandler {
	productRepo := repositories.NewProductRepository(db)
	barcodeRepo := repositories.NewBarcodeRepository(productRepo)

	return &BarcodeHandler{
		repo: barcodeRepo,
	}
}

func (h *BarcodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	var barcodeReq requests.BarcodeRequest
	if err = json.NewDecoder(r.Body).Decode(&barcodeReq); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var barcodesResponse map[int]interface{}

	if len(barcodeReq.ProductIDs) == 0 {
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

	err = json.NewEncoder(w).Encode(barcodesResponse)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
