package models

type ProductCardsLimit struct {
	FreeLimits int `json:"freeLimits"`
	PaidLimits int `json:"paidLimits"`
}
