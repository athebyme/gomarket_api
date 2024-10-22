package response

type Charc struct {
	Id    int    `json:"id"`    // ID характеристики
	Name  string `json:"name"`  // Название характеристики
	Value any    `json:"value"` // Значение характеристики. Тип значения зависит от типа характеристики
}
