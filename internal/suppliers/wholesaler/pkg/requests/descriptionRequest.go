package requests

type DescriptionRequest struct {
	FilterRequest
	IncludeEmptyDescriptions bool `json:"includeEmpty"`
}
