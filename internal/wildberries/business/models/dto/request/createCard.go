package request

import (
	"errors"
	response2 "gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type CreateCardRequestData struct {
	Brand           string                     `json:"brand"`
	Title           string                     `json:"title"`
	Description     string                     `json:"description"`
	VendorCode      string                     `json:"vendorCode"`
	Dimensions      response2.DimensionWrapper `json:"dimensions"`
	Sizes           []response2.SizeWrapper    `json:"sizes"`
	Characteristics []response2.CharcWrapper   `json:"characteristics"`
}

type CreateCardRequestWrapper struct {
	SubjectID int                     `json:"subjectID"`
	Variants  []CreateCardRequestData `json:"variants"`
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
	if req.VendorCode == "" {
		return errors.New("vendorCode is required")
	}
	if req.Dimensions.Width <= 0 || req.Dimensions.Height <= 0 || req.Dimensions.Length <= 0 {
		return errors.New("dimensions is required")
	}

	for _, v := range req.Sizes {
		if v.Price <= 0 {
			return errors.New("price is required")
		}
	}

	return nil
}
