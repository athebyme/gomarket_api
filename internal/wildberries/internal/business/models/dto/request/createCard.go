package request

import (
	"errors"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
)

type CreateCardRequestData struct {
	Brand           string                    `json:"brand"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	VendorCode      string                    `json:"vendorCode"`
	Dimensions      response.DimensionWrapper `json:"dimensions"`
	Sizes           []response.Size           `json:"sizes"`
	Characteristics []response.CharcWrapper   `json:"characteristics"`
}

type CreateCardResponseWrapper struct {
	Variants  []Variants `json:"variants"`
	SubjectID int        `json:"subjectId"`
}

func (req *CreateCardRequestData) Validate() error {
	if req.Brand == "" {
		return errors.New("brand is required")
	}
	if req.Title == "" {
		return errors.New("title is required")
	}
	if req.Description == "" {
		return errors.New("description is required")
	}
	if len(req.Sizes) == 0 {
		return errors.New("sizes is required")
	}
	if req.VendorCode == "" {
		return errors.New("vendorCode is required")
	}
	if req.Dimensions.Width <= 0 || req.Dimensions.Height <= 0 || req.Dimensions.Length <= 0 {
		return errors.New("dimensions is required")
	}

	return nil
}
