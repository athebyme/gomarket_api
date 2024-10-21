package models

type Description struct {
	ID                 int    `json:"global_id"`
	ProductDescription string `json:"product_description"`
	ProductAppellation string `json:"product_appellation"`
}
