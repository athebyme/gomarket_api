package models

import (
	"encoding/json"
	"fmt"
	"gomarketplace_api/internal/wildberries/business/models/dto/request"
)

type CompositeModel struct {
	NmID        int      `json:"nmId"`
	Media       []string `json:"media,omitempty"`
	Brand       string   `json:"brand,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
}

type MediaModel struct {
	NmID int      `json:"nmId"`
	URLs []string `json:"data"`
}

type BrandModel struct {
	NmID  int    `json:"nmId"`
	Brand string `json:"brand"`
}

type AppellationModel struct {
	NmID        int    `json:"nmId"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type SequentialModel struct {
	Models []request.Model
}

func (m CompositeModel) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m MediaModel) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m BrandModel) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m AppellationModel) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m SequentialModel) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

// Merge combines another model's data into CompositeModel
func (m *CompositeModel) Merge(model request.Model) error {
	switch v := model.(type) {
	case *MediaModel:
		m.NmID = v.NmID
		m.Media = v.URLs
	case *BrandModel:
		m.NmID = v.NmID
		m.Brand = v.Brand
	case *AppellationModel:
		m.NmID = v.NmID
		m.Title = v.Title
		m.Description = v.Description
	case *CompositeModel:
		return m.mergeComposite(v)
	default:
		return fmt.Errorf("unsupported model type for merge: %T", model)
	}
	return nil
}

func (m *CompositeModel) mergeComposite(other *CompositeModel) error {
	if m.NmID == 0 {
		m.NmID = other.NmID
	} else if m.NmID != other.NmID {
		return fmt.Errorf("nmId mismatch: %d != %d", m.NmID, other.NmID)
	}

	if len(other.Media) > 0 {
		m.Media = other.Media
	}
	if other.Brand != "" {
		m.Brand = other.Brand
	}
	if other.Title != "" {
		m.Title = other.Title
	}
	if other.Description != "" {
		m.Description = other.Description
	}
	return nil
}
