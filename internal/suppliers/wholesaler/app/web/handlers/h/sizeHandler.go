package h

import (
	"database/sql"
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/business"
	"gomarketplace_api/internal/suppliers/wholesaler/models"
	"gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	"gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"
	"io"
	"net/http"
)

type SizeHandler struct {
	service *business.SizeService
}

func NewSizeHandler(db *sql.DB, writer io.Writer) *SizeHandler {
	sizeRepo := repositories.NewSizeRepository(db, writer)
	sizeService := business.NewSizeService(sizeRepo, writer)

	return &SizeHandler{
		service: sizeService,
	}
}

func (h *SizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sizes map[int][]models.SizeWrapper
	var err error

	var sizeReq requests.SizeRequest
	if err := json.NewDecoder(r.Body).Decode(&sizeReq); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	if len(sizeReq.ProductIDs) == 0 {
		sizes, err = h.service.GetSizes()
	} else {
		sizes, err = h.service.GetSizesByIDs(sizeReq.ProductIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	err = json.NewEncoder(w).Encode(sizes)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
