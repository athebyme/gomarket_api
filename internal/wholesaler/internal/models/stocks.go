package models

type Stocks struct {
	ID            int    `json:"global_id"`
	MainArticular string `json:"main_articular"`
	Stocks        int    `json:"stocks"`
}
