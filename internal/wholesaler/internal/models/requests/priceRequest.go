package requests

type PriceRequest struct {
	ProductIDs []int `json:"productIDs"`
	All        bool  `json:"all"`
}
