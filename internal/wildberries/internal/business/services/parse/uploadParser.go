package parse

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/services/builder"
)

type CreateEngine struct {
}

func NewUploadParser() *CreateEngine {
	return &CreateEngine{}
}

func (e *CreateEngine) Fetch(data interface{}) (interface{}, error) {
	jsonData, ok := data.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected []byte")
	}

	// пробуем как одиночный объект
	var singleRequest request.CreateCardRequestData
	if err := json.Unmarshal(jsonData, &singleRequest); err == nil {
		cardBuilder := builder.CreateCardBuilder{}
		cardBuilder.WithBrand(singleRequest.Brand).
			WithTitle(singleRequest.Title).
			WithDescription(singleRequest.Description).
			WithVendorCode(singleRequest.VendorCode).
			WithDimensionWrapper(singleRequest.Dimensions).
			WithSizes(singleRequest.Sizes).
			WithCharacteristics(singleRequest.Characteristics)

		return cardBuilder.Build()
	}

	// если это не одиночный объект, пробуем массив
	var requestArray []request.CreateCardRequestData
	if err := json.Unmarshal(jsonData, &requestArray); err != nil {
		return nil, fmt.Errorf("failed to parse input data: %w", err)
	}

	// строим массив
	var results []interface{}
	for _, req := range requestArray {
		cardBuilder := builder.CreateCardBuilder{}
		cardBuilder.WithBrand(req.Brand).
			WithTitle(req.Title).
			WithDescription(req.Description).
			WithVendorCode(req.VendorCode).
			WithDimensionWrapper(req.Dimensions).
			WithSizes(req.Sizes).
			WithCharacteristics(req.Characteristics)

		card, err := cardBuilder.Build()
		if err != nil {
			return nil, fmt.Errorf("error building card: %w", err)
		}
		results = append(results, card)
	}

	return results, nil
}
