package response

type Nomenclature struct {
	NmID            int        `json:"nmID"`
	ImtID           int        `json:"imtID"`
	NmUUID          string     `json:"nmUUID"`
	SubjectID       int        `json:"subjectID"`
	VendorCode      string     `json:"vendorCode"`
	SubjectName     string     `json:"subjectName"`
	Brand           string     `json:"brand"`
	Title           string     `json:"title"`
	Photos          []Photo    `json:"photos"`
	Video           string     `json:"video"`
	Dimensions      Dimensions `json:"dimensions"`
	Characteristics []Charc    `json:"characteristics"`
	Sizes           []Size     `json:"sizes"`
	Tags            []Tag      `json:"tags"`
	CreatedAt       string     `json:"createdAt"`
	UpdatedAt       string     `json:"updatedAt"`
}
