package h

import (
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	"gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"
	"net/http"
)

type BrandHandler struct {
	repo *repositories.BrandRepository
}

func NewBrandHandler(repo *repositories.BrandRepository) *BrandHandler {
	return &BrandHandler{
		repo: repo,
	}
}

func (h *BrandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	var req requests.BrandRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var brands map[int]string

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

	err = json.NewEncoder(w).Encode(brands)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
