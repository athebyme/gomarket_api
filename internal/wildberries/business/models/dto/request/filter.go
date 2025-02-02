package request

type Filter struct {
	/*
		Фильтр по фото:
		0 — только карточки без фото
		1 — только карточки с фото
		-1 — все карточки товара
	*/
	WithPhoto int `json:"withPhoto"`

	/*
		Поиск по артикулу продавца, артикулу WB, баркоду
	*/
	TextSearch string `json:"textSearch"`

	/*
		Поиск по ID тегов
	*/
	TagIDs []int `json:"tagIDs"`

	/*
		Фильтр по категории. true - только разрешённые, false - все. Не используется в песочнице.
	*/
	AllowedCategoriesOnly bool `json:"allowedCategoriesOnly"`

	/*
		Поиск по id предметов
	*/
	ObjectIDs []int `json:"objectIDs"`

	/*
		Поиск по брендам
	*/
	Brands []string `json:"brands"`

	/*
		Поиск по ID карточки товара
	*/
	ImtID int `json:"imtID"`
}
