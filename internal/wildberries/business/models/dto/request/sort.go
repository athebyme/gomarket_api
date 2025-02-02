package request

type Sort struct {
	Ascending bool `json:"ascending"` // Сортировать по полю updatedAt (false - по убыванию, true - по возрастанию)
}
