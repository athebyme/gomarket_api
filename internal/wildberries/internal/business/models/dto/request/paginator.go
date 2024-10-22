package request

type Paginator struct {
	UpdatedAt string `json:"updatedAt"`
	NmID      int    `json:"nmId"`
	Total     int    `json:"total"`
}
