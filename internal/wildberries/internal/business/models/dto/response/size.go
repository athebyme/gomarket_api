package response

type Size struct {
	ChrtID   int      `json:"chrtID"`   //Числовой ID размера для данного артикула WB
	TechSize string   `json:"techSize"` // Размер товара (А, XXL, 57 и др.)
	WbSize   string   `json:"wbSize"`   // Российский размер товара
	Skus     []string `json:"skus"`     // Баркод товара
}

type SizeWrapper struct {
	TechSize string   `json:"techSize"` // Размер товара (А, XXL, 57 и др.)
	WbSize   string   `json:"wbSize"`   // Российский размер товара
	Skus     []string `json:"skus"`     // Баркод товара
	Price    int      `json:"price"`
}

func (s *Size) Wrap() SizeWrapper {
	return SizeWrapper{
		TechSize: s.TechSize,
		WbSize:   s.WbSize,
		Skus:     s.Skus,
		Price:    0,
	}
}
