package service

import "database/sql"

type PostgresLoader struct {
	db *sql.DB
}

func NewPostgresLoader(db *sql.DB) *PostgresLoader {
	return &PostgresLoader{db: db}
}

func (p *PostgresLoader) InsertCardActual(cardData map[string]interface{}) error {
	_, err := p.db.Exec(`
		INSERT INTO wildberries.cards_actual (global_id, nm_id, vendor_code, version, version_data)
		VALUES ($1, $2, $3, $4, $5)`,
		cardData["globalID"], cardData["nmID"], cardData["vendorCode"],
		cardData["version"], cardData["version_data"])
	return err
}

func (p *PostgresLoader) InsertCardHistory(cardData map[string]interface{}, version int) error {
	_, err := p.db.Exec(`
		INSERT INTO wildberries.cards_history (global_id, nm_id, vendor_code, vesion, version_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		cardData["globalID"], cardData["nmID"], cardData["vendorCode"],
		cardData["version"], cardData["version_data"])
	return err
}

func (p *PostgresLoader) RollbackCard(nmID int, versionID int) error {
	_, err := p.db.Exec(`
        INSERT INTO wildberries.cards_history (global_id, nm_id, vendor_code, version, version_data, created_at)
        SELECT global_id, nm_id, vendor_code, version, version_data, NOW()
        FROM wildberries.cards_actual 
        WHERE nm_id = $1
    `, nmID)
	if err != nil {
		return err
	}

	// Шаг 2: Получаем данные версии, на которую нужно откатиться
	var versionData []byte
	err = p.db.QueryRow(`
        SELECT version_data 
        FROM wildberries.cards_history 
        WHERE nm_id = $1 AND version_id = $2
    `, nmID, versionID).Scan(&versionData)
	if err != nil {
		return err
	}

	// Шаг 3: Обновляем основную запись в cards_actual данными из версии
	_, err = p.db.Exec(`
        UPDATE wildberries.cards_actual
        SET version = version + 1, version_data = $1, updated_at = NOW()
        WHERE nm_id = $2
    `, versionData, nmID)
	return err
}

func (p *PostgresLoader) SaveVersion(cardData map[string]interface{}) error {
	// Получаем текущую версию из таблицы wildberries.cards_actual
	var currentVersion int
	err := p.db.QueryRow(`
        SELECT version 
        FROM wildberries.cards_actual 
        WHERE nm_id = $1
    `, cardData["nmID"]).Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Сохраняем текущую версию в history
	_, err = p.db.Exec(`
        INSERT INTO wildberries.cards_history (global_id, nm_id, vendor_code, version, version_data)
        SELECT global_id, nm_id, vendor_code, version, version_data
        FROM wildberries.cards_actual 
        WHERE nm_id = $1
    `, cardData["nmID"])
	if err != nil {
		return err
	}

	// Увеличиваем версию на 1 и обновляем запись в cards_actual новыми данными
	newVersion := currentVersion + 1
	_, err = p.db.Exec(`
        UPDATE wildberries.cards_actual
        SET version = $1, version_data = $2, updated_at = NOW()
        WHERE nm_id = $3
    `, newVersion, cardData, cardData["nmID"])
	return err
}
