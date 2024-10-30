package service

import "database/sql"

type DatabaseLoader interface {
	InsertCardActual(cardData map[string]interface{}) error
	InsertCardHistory(cardData map[string]interface{}, version int) error
	RollbackCard(cardData map[string]interface{}, version int) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) RowScanner
}

type RowScanner interface {
	Scan(dest ...interface{}) error
}

type Row struct {
	row *sql.Row
}

func (r *Row) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}
