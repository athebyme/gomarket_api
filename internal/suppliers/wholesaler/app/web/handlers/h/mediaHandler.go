package h

import (
	"encoding/json"
	"gomarketplace_api/internal/suppliers/wholesaler/pkg/requests"
	"gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"
	"log"
	"net/http"
	"time"
)

type MediaHandler struct {
	repo *repositories.MediaRepository
}

func NewMediaHandler(repo *repositories.MediaRepository) *MediaHandler {
	return &MediaHandler{
		repo: repo,
	}
}

func (h *MediaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var mediaReq requests.MediaRequest
	if err := json.NewDecoder(r.Body).Decode(&mediaReq); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var mediaMap map[int][]string
	var err error

	startTime := time.Now()
	if len(mediaReq.ProductIDs) == 0 {
		mediaMap, err = h.repo.GetMediaSources(mediaReq.Censored)
		if err != nil {
			http.Error(w, "Failed to fetch all media sources", http.StatusInternalServerError)
			return
		}
	} else {
		mediaMap, err = h.repo.GetMediaSourcesByProductIDs(mediaReq.ProductIDs, mediaReq.Censored, mediaReq.ImageSize)
		if err != nil {
			http.Error(w, "Failed to fetch media sources", http.StatusInternalServerError)
			return
		}
	}
	log.Printf("media handler response execution time: %v", time.Since(startTime))

	err = json.NewEncoder(w).Encode(mediaMap)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
