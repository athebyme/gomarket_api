package requests

type PriceRequest struct {
	FilterRequest
	All bool `json:"all"`
}
