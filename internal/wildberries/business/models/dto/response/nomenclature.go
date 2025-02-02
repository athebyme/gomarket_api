package response

import (
	"fmt"
	"regexp"
	"strconv"
)

type Nomenclature struct {
	NmID            int        `json:"nmID"`
	ImtID           int        `json:"imtID"`
	NmUUID          string     `json:"nmUUID"`
	SubjectID       int        `json:"subjectID"`
	VendorCode      string     `json:"vendorCode"`
	SubjectName     string     `json:"subjectName"`
	Brand           string     `json:"brand"`
	Title           string     `json:"title"`
	Photos          []Photo    `json:"photos"`
	Video           string     `json:"video"`
	Dimensions      Dimensions `json:"dimensions"`
	Characteristics []Charc    `json:"characteristics"`
	Sizes           []Size     `json:"sizes"`
	Tags            []Tag      `json:"tags"`
	CreatedAt       string     `json:"createdAt"`
	UpdatedAt       string     `json:"updatedAt"`
}

func (n *Nomenclature) GlobalID() (int, error) {
	pattern := `\w*-(\d*)-\w*`
	re := regexp.MustCompile(pattern)
	match := re.FindAllStringSubmatch(n.VendorCode, -1)

	// Проверяем, что найдено хотя бы одно совпадение и нужная группа
	if len(match) == 0 || len(match[0]) < 2 {
		return 0, fmt.Errorf("no match found in VendorCode: %s", n.VendorCode)
	}

	// Преобразуем найденное значение в int
	globalID, err := strconv.Atoi(match[0][1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert global ID: %w", err)
	}

	return globalID, nil
}
