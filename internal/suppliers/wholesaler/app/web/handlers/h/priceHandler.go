package h

import (
	"database/sql"
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/business"
	"gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	"gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"
	"net/http"
)

type PriceHandler struct {
	business.PriceService
}

func NewPriceHandler(db *sql.DB) *PriceHandler {
	priceRepo := repositories.NewPriceRepository(db)
	priceService := business.NewPriceEngine(priceRepo)

	return &PriceHandler{
		PriceService: priceService,
	}
}

func (h *PriceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	var priceRequest requests.PriceRequest
	if err = json.NewDecoder(r.Body).Decode(&priceRequest); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var response interface{}

	if len(priceRequest.ProductIDs) == 0 {
		response, err = h.GetPrices(priceRequest.All)
		if err != nil {
			http.Error(w, "Failed to fetch all media sources", http.StatusInternalServerError)
			return
		}
	} else {
		response, err = h.GetPricesById(priceRequest.ProductIDs)
		if err != nil {
			http.Error(w, "Failed to fetch media sources", http.StatusInternalServerError)
			return
		}
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
