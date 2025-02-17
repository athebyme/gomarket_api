package csv_to_postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
	"strings"
)

type PostgresUpdater struct {
	DB        *sql.DB
	Schema    string
	TableName string
	Columns   []string
}

func NewPostgresUpdater(db *sql.DB, schema, tableName string, columns []string) *PostgresUpdater {
	return &PostgresUpdater{
		DB:        db,
		Schema:    schema,
		TableName: tableName,
		Columns:   columns,
	}
}
func (u *PostgresUpdater) SetNewColumnNaming(columns []string) *PostgresUpdater {
	if len(columns) == 0 {
		return u
	}
	u.Columns = columns
	return u
}

func (u *PostgresUpdater) SetNewSchema(schema string) *PostgresUpdater {
	if schema == "" {
		return u
	}
	u.Schema = schema
	return u
}

func (u *PostgresUpdater) SetNewTableName(tableName string) *PostgresUpdater {
	if tableName == "" {
		return u
	}
	u.TableName = tableName
	return u
}

// UpdateData принимает подготовленные CSV данные и выполняет обновление через транзакцию.
func (u *PostgresUpdater) UpdateData(csvData [][]interface{}, ctx context.Context) error {
	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tempTableName := "temp_" + u.TableName
	createTempTableQuery := fmt.Sprintf(`
        CREATE TEMP TABLE %s AS
        SELECT * FROM %s.%s WHERE 1=0
    `, tempTableName, u.Schema, u.TableName)
	if _, err := tx.ExecContext(ctx, createTempTableQuery); err != nil {
		return fmt.Errorf("create temp table error: %w", err)
	}
	log.Printf("Temp table %s создан", tempTableName)

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn(tempTableName, u.Columns...))
	if err != nil {
		return fmt.Errorf("prepare copyin error: %w", err)
	}

	for i, row := range csvData[1:] {
		log.Printf("Processing row %d", i)
		log.Printf("Row data: %v", row)

		if _, err := stmt.ExecContext(ctx, row...); err != nil {
			log.Printf("Error in row %d: %v", i, err)
			return fmt.Errorf("exec copyin error at row %d: %w", i, err)
		}
	}
	if _, err = stmt.ExecContext(ctx); err != nil {
		return fmt.Errorf("final exec copyin error: %w", err)
	}
	if err = stmt.Close(); err != nil {
		return fmt.Errorf("close stmt error: %w", err)
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO %s.%s (%s)
		SELECT %s 
		FROM %s AS temp
		LEFT JOIN %s.%s AS main ON temp.%s = main.%s
		WHERE main.%s IS NULL
		ON CONFLICT (%s) DO NOTHING
	`, u.Schema, u.TableName,
		strings.Join(u.Columns, ","),
		strings.Join(u.prefixedColumns("temp."), ","),
		tempTableName,
		u.Schema, u.TableName,
		u.Columns[0], u.Columns[0],
		u.Columns[0],
		u.Columns[0])
	log.Printf("Выполнение запроса: %s", insertQuery)

	if _, err = tx.ExecContext(ctx, insertQuery); err != nil {
		return fmt.Errorf("insert execution error: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}

	return nil
}

func (u *PostgresUpdater) prefixedColumns(prefix string) []string {
	cols := make([]string, len(u.Columns))
	for i, col := range u.Columns {
		cols[i] = prefix + col
	}
	return cols
}
