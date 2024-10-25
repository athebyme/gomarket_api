package response

type Category struct {
	SubjectID   int    `json:"subjectID"`
	ParentID    int    `json:"parentID"`
	SubjectName string `json:"subjectName"`
	ParentName  string `json:"parentName"`
}
