package request

type Pagination interface {
}

type Paginator struct {
	UpdatedAt string `json:"updatedAt"`
	NmID      int    `json:"nmId"`
	Total     int    `json:"total"`
}

type PaginatorCursor struct {
	UpdatedAt string `json:"updatedAt"`
	NmID      int    `json:"nmId"`
}

func (p *Paginator) GetPaginatorCursor() PaginatorCursor {
	return PaginatorCursor{UpdatedAt: p.UpdatedAt, NmID: p.NmID}
}
