package request

type Pagination interface {
	GetPaginatorCursor() PaginatorCursor
	TotalCards() int
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

func (p *Paginator) TotalCards() int {
	return p.Total
}

func (pc *PaginatorCursor) TotalCards() int {
	return -1
}

func (pc *PaginatorCursor) GetPaginatorCursor() PaginatorCursor {
	return *pc
}
