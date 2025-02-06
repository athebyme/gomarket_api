package requests

import "gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"

type MediaRequest struct {
	FilterRequest
	Censored  bool                   `json:"censored"`
	ImageSize repositories.ImageSize `json:"imageSize"`
}
