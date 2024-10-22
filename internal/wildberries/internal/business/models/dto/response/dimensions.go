package response

type Dimensions struct {
	Length  int  `json:"length"`
	Width   int  `json:"width"`
	Height  int  `json:"height"`
	IsValid bool `json:"isValid"`
}
