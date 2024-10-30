package response

type Dimensions struct {
	Length  int  `json:"length"`
	Width   int  `json:"width"`
	Height  int  `json:"height"`
	IsValid bool `json:"isValid"`
}

type DimensionWrapper struct {
	Length int `json:"length"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (d *Dimensions) Unwrap() *DimensionWrapper {
	return &DimensionWrapper{
		Length: d.Length,
		Width:  d.Width,
		Height: d.Height,
	}
}
