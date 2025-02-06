package csv_to_postgres

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"log"
)

// Processor отвечает за чтение и фильтрацию CSV данных.
type Processor struct {
	Columns []string
}

// NewProcessor создаёт новый Processor.
func NewProcessor(columns []string) *Processor {
	return &Processor{Columns: columns}
}

func (p *Processor) SetNewColumnNaming(columns []string) *Processor {
	if len(columns) == 0 {
		return p
	}
	p.Columns = columns
	return p
}

// ProcessCSV читает CSV данные из reader, декодируя из Windows-1251, и возвращает двумерный срез строк.
func (p *Processor) ProcessCSV(reader io.Reader, renaming []string) ([][]string, error) {
	decoder := transform.NewReader(reader, charmap.Windows1251.NewDecoder())
	csvReader := csv.NewReader(decoder)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1

	allRows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv read error: %w", err)
	}
	if len(allRows) == 0 {
		return nil, fmt.Errorf("csv data is empty")
	}

	var header []string
	var data [][]string

	if len(allRows) > 0 && p.isHeader(allRows[0]) {
		header = allRows[0]
		data = allRows[1:]
	} else {
		header = p.Columns
		data = allRows
	}

	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	filteredRows := [][]string{p.Columns}
	for _, row := range data {
		filteredRow := make([]string, len(p.Columns))
		for i, col := range p.Columns {
			if idx, ok := columnMap[col]; ok && idx < len(row) {
				filteredRow[i] = row[idx]
			} else {
				filteredRow[i] = ""
			}
		}
		filteredRows = append(filteredRows, filteredRow)
	}

	if renaming != nil && len(renaming) == len(p.Columns) {
		for i := range filteredRows[0] {
			log.Printf("Переименование колонки [%s] в [%s]", filteredRows[0][i], renaming[i])
			filteredRows[0][i] = renaming[i]
		}
	}

	return filteredRows, nil
}

func (p *Processor) isHeader(row []string) bool {
	for _, col := range p.Columns {
		if indexOf(row, col) >= 0 {
			return true
		}
	}
	return false
}

func indexOf(slice []string, str string) int {
	for i, s := range slice {
		if s == str {
			return i
		}
	}
	return -1
}
