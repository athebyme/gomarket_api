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
	Sizes           response.SizeWrapper      `json:"sizes"`
	Characteristics []response.CharcWrapper   `json:"characteristics"`
}

type CreateCardRequestWrapper struct {
	Variants  []CreateCardRequestData `json:"variants"`
	SubjectID int                     `json:"subjectId"`
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
	if req.Sizes.Price == 0 {
		return errors.New("price is required")
	}
	if req.VendorCode == "" {
		return errors.New("vendorCode is required")
	}
	if req.Dimensions.Width <= 0 || req.Dimensions.Height <= 0 || req.Dimensions.Length <= 0 {
		return errors.New("dimensions is required")
	}

	return nil
}
