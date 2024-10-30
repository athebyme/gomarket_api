package service

import "database/sql"

type PostgresLoader struct {
	db *sql.DB
}

func NewPostgresLoader(db *sql.DB) *PostgresLoader {
	return &PostgresLoader{db: db}
}

func (p *PostgresLoader) InsertCardActual(cardData map[string]interface{}) error {
	// Логика вставки данных в таблицу cards_actual
	_, err := p.db.Exec(`
		INSERT INTO wildberries.cards_actual (global_id, nm_id, vendor_code, title, description, brand, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		cardData["globalID"], cardData["nmID"], cardData["vendorCode"],
		cardData["title"], cardData["description"], cardData["brand"],
		cardData["createdAt"], cardData["updatedAt"])
	return err
}

func (p *PostgresLoader) InsertCardHistory(cardData map[string]interface{}, version int) error {
	// Логика вставки данных в таблицу cards_history
	_, err := p.db.Exec(`
		INSERT INTO wildberries.cards_history (global_id, version, nm_id, vendor_code, title, description, brand, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		cardData["globalID"], version, cardData["nmID"], cardData["vendorCode"],
		cardData["title"], cardData["description"], cardData["brand"],
		cardData["createdAt"], cardData["updatedAt"])
	return err
}

func (p *PostgresLoader) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

func (p *PostgresLoader) QueryRow(query string, args ...interface{}) RowScanner {
	return &Row{row: p.db.QueryRow(query, args...)}
}
