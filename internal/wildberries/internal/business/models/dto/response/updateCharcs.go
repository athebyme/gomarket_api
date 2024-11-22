package response

type Charc struct {
	Id    int    `json:"id"`    // ID характеристики
	Name  string `json:"name"`  // Название характеристики
	Value any    `json:"value"` // Значение характеристики. Тип значения зависит от типа характеристики
}

type CharcWrapper struct {
	Id    int `json:"id"`    // ID характеристики
	Value any `json:"value"` // Значение характеристики. Тип значения зависит от типа характеристики
}

func (c *Charc) Unwrap() *CharcWrapper {
	return &CharcWrapper{
		Id:    c.Id,
		Value: c.Value,
	}
}
