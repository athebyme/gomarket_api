package converters

import (
	"strconv"
	"strings"
)

type ColumnConverter func(string) (interface{}, error)

func DecimalConverter(cell string) (interface{}, error) {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return nil, nil
	}
	return strconv.ParseFloat(cell, 64)
}

func IntConverter(cell string) (interface{}, error) {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return nil, nil
	}
	return strconv.Atoi(cell)
}

func DefaultConverter(cell string) (interface{}, error) {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return nil, nil
	}
	return cell, nil
}
