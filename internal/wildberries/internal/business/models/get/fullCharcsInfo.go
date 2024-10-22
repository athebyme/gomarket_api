package get

type FullCharcsInfo struct {
	CharcID          int      `json:"charcID"`
	SubjectName      string   `json:"subjectName"`
	SubjectID        int      `json:"subjectID"`
	Name             string   `json:"name"`
	Required         bool     `json:"required"`
	UnitName         string   `json:"unitName"`
	MaxCount         int      `json:"maxCount"`
	Popular          bool     `json:"popular"`
	CharcType        int      `json:"charcType"`
	Error            bool     `json:"error"`
	ErrorText        string   `json:"errorText"`
	AdditionalErrors []string `json:"additionalErrors"`
}
