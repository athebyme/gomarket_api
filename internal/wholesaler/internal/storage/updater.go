package storage

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Updater interface {
	Update(args ...[]string) error
}

type DataSource struct {
	InfURL           string
	CSVURL           string
	LastUpdateColumn string
}

type PostgresUpdater struct {
	DB         *sql.DB
	Schema     string
	TableName  string
	Columns    []string
	LastModCol string
	InfURL     string
	CSVURL     string
}

func NewPostgresUpdater(db *sql.DB, schema string, tableName string, columns []string, lastModCol, infURL, csvURL string) *PostgresUpdater {
	return &PostgresUpdater{
		DB:         db,
		Schema:     schema,
		TableName:  tableName,
		Columns:    columns,
		LastModCol: lastModCol,
		InfURL:     infURL,
		CSVURL:     csvURL,
	}
}
func (pu *PostgresUpdater) fetchInfTime() (time.Time, error) {
	resp, err := http.Get(pu.InfURL)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return time.Time{}, errors.New("empty response from inf file")
	}
	modTimeStr := scanner.Text()

	return time.Parse("2006-01-02 15:04:05", modTimeStr)
}
func (pu *PostgresUpdater) fetchCSVData(renaming []string) ([][]string, error) {
	resp, err := http.Get(pu.CSVURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Fetched CSV file size: %d bytes", len(body))

	// Use Windows-1251 encoding directly with charmap
	reader := transform.NewReader(bytes.NewReader(body), charmap.Windows1251.NewDecoder())

	// Ensure the CSV reader handles the correct delimiter and quoted fields correctly
	r := csv.NewReader(reader)
	r.Comma = ';'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1 // Allow variable number of fields per record

	allRows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	log.Printf("Number of rows read: %d", len(allRows))

	var header []string
	var data [][]string

	// Check if the first row contains the header by checking the first value
	if len(allRows) > 0 && isHeader(allRows[0], pu.Columns) {
		header = allRows[0]
		data = allRows[1:]
	} else {
		header = pu.Columns
		data = allRows
	}

	// Create a map for quick lookup of desired columns
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	// Initialize filteredRows with the provided Columns as header
	filteredRows := [][]string{pu.Columns}
	for _, row := range data {
		filteredRow := make([]string, len(pu.Columns))
		for i, col := range pu.Columns {
			if index, found := columnMap[col]; found && index < len(row) {
				filteredRow[i] = row[index]
			} else {
				// If the column is not in the original data, set it to an empty string
				filteredRow[i] = ""
			}
		}
		filteredRows = append(filteredRows, filteredRow)
	}

	if renaming != nil {
		for i, v := range renaming {
			filteredRows[0][i] = v
		}
	}

	return filteredRows, nil
}

func isHeader(row, columns []string) bool {
	for _, col := range columns {
		if indexOf(row, col) >= 0 {
			return true
		}
	}
	return false
}

// Вспомогательная функция для поиска индекса элемента в срезе
func indexOf(slice []string, str string) int {
	for i, s := range slice {
		if s == str {
			return i
		}
	}
	return -1
}

func (pu *PostgresUpdater) updateDatabase(csvData [][]string) error {
	tx, err := pu.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create a temporary table without specifying the schema
	tempTableName := "temp_" + pu.TableName
	_, err = tx.Exec(fmt.Sprintf(`
        CREATE TEMP TABLE %s AS
        SELECT * FROM %s.%s WHERE 1=0
    `, tempTableName, pu.Schema, pu.TableName))
	if err != nil {
		return err
	}

	// Prepare the COPY command to insert data into the temporary table
	stmt, err := tx.Prepare(pq.CopyIn(tempTableName, pu.Columns...))
	if err != nil {
		return err
	}

	for _, row := range csvData[1:] {
		_, err = stmt.Exec(convertRowToInterfaceSlice(row)...)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	// Insert data from the temporary table into the main table, checking for duplicates
	insertQuery := fmt.Sprintf(`
        INSERT INTO %s.%s (%s)
        SELECT %s FROM %s
        WHERE NOT EXISTS (
            SELECT 1 FROM %s.%s WHERE %s.%s = %s.%s
        )
    `, pu.Schema, pu.TableName, strings.Join(pu.Columns, ","), strings.Join(pu.Columns, ","), tempTableName, pu.Schema, pu.TableName, pu.TableName, pu.Columns[0], tempTableName, pu.Columns[0])

	_, err = tx.Exec(insertQuery)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

func convertRowToInterfaceSlice(row []string) []any {
	result := make([]any, len(row))
	for i, v := range row {
		result[i] = v
	}
	return result
}

func (pu *PostgresUpdater) Update(args ...[]string) error {
	modTime, err := pu.fetchInfTime()
	if err != nil {
		return err
	}

	var storedTime time.Time
	err = pu.DB.QueryRow(
		"SELECT last_update FROM metadata WHERE key_name = $1",
		pu.LastModCol,
	).Scan(&storedTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Если данных нет, устанавливаем начальное значение
			storedTime = time.Time{}
		} else {
			return err
		}
	}

	if modTime.After(storedTime) {
		log.Printf("Updating data from %s...", pu.CSVURL)

		var renamedCols []string
		if len(args) > 0 {
			renamedCols = args[0]
		}

		csvData, err := pu.fetchCSVData(renamedCols)
		if err != nil {
			return err
		}

		if err := pu.updateDatabase(csvData); err != nil {
			return err
		}

		// Обновляем или вставляем время последнего обновления
		_, err = pu.DB.Exec(`
			INSERT INTO metadata (key_name, value, last_update)
			VALUES ($1, $2, $3)
			ON CONFLICT (key_name) DO UPDATE SET last_update = EXCLUDED.last_update
		`, pu.LastModCol, pu.LastModCol, modTime)
		if err != nil {
			return err
		}

		log.Printf("Data from %s updated successfully.", pu.CSVURL)
	} else {
		log.Printf("No update necessary for %s.", pu.CSVURL)
	}

	return nil
}