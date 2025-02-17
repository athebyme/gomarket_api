package csv_to_postgres

import (
	"encoding/csv"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"gomarketplace_api/pkg/business/service/csv_to_postgres/converters"
	"io"
	"log"
)

// Processor отвечает за чтение и фильтрацию CSV данных.
type Processor struct {
	columns          []string
	columnConverters map[string]converters.ColumnConverter
}

// NewProcessor создаёт новый Processor.
func NewProcessor(columns []string) *Processor {
	return &Processor{
		columns:          columns,
		columnConverters: map[string]converters.ColumnConverter{},
	}
}

func (p *Processor) SetNewColumnNaming(columns []string) *Processor {
	if len(columns) == 0 {
		return p
	}
	p.columns = columns
	return p
}

func (p *Processor) SetNewConverters(converters map[string]converters.ColumnConverter) *Processor {
	if len(converters) == 0 {
		return p
	}
	p.columnConverters = converters
	return p
}

// ProcessCSV читает CSV данные из reader, декодируя из Windows-1251, и возвращает двумерный срез строк.
func (p *Processor) ProcessCSV(reader io.Reader, renaming []string) ([][]interface{}, []string, error) {
	decoder := transform.NewReader(reader, charmap.Windows1251.NewDecoder())
	csvReader := csv.NewReader(decoder)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1

	allRows, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("csv read error: %w", err)
	}
	if len(allRows) == 0 {
		return nil, nil, fmt.Errorf("csv data is empty")
	}

	var header []string
	var data [][]string

	if len(allRows) > 0 && p.isHeader(allRows[0]) {
		header = allRows[0]
		data = allRows[1:]
	} else {
		header = p.columns
		data = allRows
	}

	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	filteredRows := [][]string{p.columns}
	for _, row := range data {
		filteredRow := make([]string, len(p.columns))
		for i, col := range p.columns {
			if idx, ok := columnMap[col]; ok && idx < len(row) {
				filteredRow[i] = row[idx]
			} else {
				filteredRow[i] = ""
			}
		}
		filteredRows = append(filteredRows, filteredRow)
	}

	if renaming != nil && len(renaming) == len(p.columns) {
		for i := range filteredRows[0] {
			log.Printf("Переименование колонки [%s] в [%s]", filteredRows[0][i], renaming[i])
			filteredRows[0][i] = renaming[i]
		}
	}

	convertedData := make([][]interface{}, len(filteredRows))
	for i, row := range filteredRows {
		if i == 0 {
			convertedData[i] = stringSliceToInterface(row)
			continue
		}

		converted, err := convertRowToInterfaceSlice(row, p.columns, p.columnConverters)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d conversion error: %w", i, err)
		}
		convertedData[i] = converted
	}

	return convertedData, filteredRows[0], nil
}

func (p *Processor) isHeader(row []string) bool {
	for _, col := range p.columns {
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

func stringSliceToInterface(src []string) []interface{} {
	res := make([]interface{}, len(src))
	for i, v := range src {
		res[i] = v
	}
	return res
}

func convertRowToInterfaceSlice(row []string, columns []string, colConverters map[string]converters.ColumnConverter) ([]interface{}, error) {
	if len(row) != len(columns) {
		return nil, fmt.Errorf(
			"количество значений (%d) не совпадает с количеством колонок (%d)",
			len(row),
			len(columns),
		)
	}

	result := make([]interface{}, len(row))

	for i, cell := range row {
		colName := columns[i]
		var val interface{}
		var err error

		if conv, exists := colConverters[colName]; exists {
			val, err = conv(cell)
		} else {
			val, err = converters.DefaultConverter(cell)
		}

		if err != nil {
			return nil, fmt.Errorf(
				"ошибка конвертации для колонки %q, значение %q: %w",
				colName,
				cell,
				err,
			)
		}

		result[i] = val
	}

	return result, nil
}
