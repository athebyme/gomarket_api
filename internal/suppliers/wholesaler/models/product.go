package models

type Product struct {
	ID             int    `json:"global_id"`
	Model          string `json:"model"`
	Appellation    string `json:"appellation"`
	Category       string `json:"category"`
	Brand          string `json:"brand"`
	Country        string `json:"country"`
	ProductType    string `json:"product_type"`
	Features       string `json:"features"`
	Sex            string `json:"sex"`
	Color          string `json:"color"`
	Dimension      string `json:"dimension"`
	Package        string `json:"package"`
	Media          string `json:"media"`
	Barcodes       string `json:"barcodes"`
	Material       string `json:"material"`
	PackageBattery string `json:"package_battery"`
}
