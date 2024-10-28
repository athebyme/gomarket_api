package request

type Cursor struct {
	Limit      int `json:"limit"` // Сколько карточек товара выдать в ответе.
	Pagination Paginator
}
