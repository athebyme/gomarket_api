package requests

type MediaRequest struct {
	FilterRequest
	Censored  bool `json:"censored"`
	ImageSize int  `json:"imageSize"`
}
