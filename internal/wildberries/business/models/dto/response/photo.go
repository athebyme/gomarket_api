package response

type Photo struct {
	Big    string `json:"big"`
	Tiny   string `json:"tm"`
	Small  string `json:"c246x328"`
	Square string `json:"square"`
	Medium string `json:"c516x688"`
}
