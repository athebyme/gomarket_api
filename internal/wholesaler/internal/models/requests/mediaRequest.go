package requests

type MediaRequest struct {
	ProductIDs []int `json:"productIDs"`
	Censored   bool  `json:"censored"`
	ImageSize  int   `json:"imageSize"`
}
