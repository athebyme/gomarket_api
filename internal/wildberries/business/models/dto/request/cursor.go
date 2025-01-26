package request

type Cursor struct {
	Limit     int    `json:"limit"` // Сколько карточек товара выдать в ответе.
	UpdatedAt string `json:"updatedAt,omitempty"`
	NmID      int    `json:"nmId,omitempty"`
}
