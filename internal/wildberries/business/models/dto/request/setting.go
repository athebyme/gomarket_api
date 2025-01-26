package request

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type SettingsRequestWrapper struct {
	Settings Settings `json:"settings"`
}

type Settings struct {
	Sort   Sort   `json:"sort"`
	Filter Filter `json:"filter"`
	Cursor Cursor `json:"cursor"`
}

func (s *Settings) CreateRequestBody() (*bytes.Buffer, error) {
	wrapper := SettingsRequestWrapper{Settings: *s}

	jsonData, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("marshalling settings: %w", err)
	}
	return bytes.NewBuffer(jsonData), nil
}
