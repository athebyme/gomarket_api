package get

import (
	"gomarketplace_api/pkg/business/service"
	"log"
	"time"
)

type UpdateService struct {
	db       *service.DatabaseLoader
	updateCh chan bool // Канал для получения сигнала об изменении
}

func (s *UpdateService) InitializeCard(cardData map[string]interface{}) error {
	query := `
		INSERT INTO wildberries.cards_history (global_id, nm_id, vendor_code, version, version_data)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.Exec(query, cardData["globalID"], cardData["nmID"], cardData["vendorCode"], 1, cardData)
	return err
}

func (s *UpdateService) SaveVersion(cardData map[string]interface{}) error {
	// Получаем текущую версию из таблицы wildberries.cards_actual
	var currentVersion int
	err := s.db.QueryRow(`
        SELECT version 
        FROM wildberries.cards_actual 
        WHERE nm_id = $1
    `, cardData["nmID"]).Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Сохраняем текущую версию в history
	_, err = s.db.Exec(`
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
	_, err = s.db.Exec(`
        UPDATE wildberries.cards_actual
        SET version = $1, version_data = $2, updated_at = NOW()
        WHERE nm_id = $3
    `, newVersion, cardData, cardData["nmID"])
	return err
}

func (s *UpdateService) RollbackToVersion(nmID int, versionID int) error {
	// Шаг 1: Сохраняем текущую версию из cards_actual в cards_history
	_, err := s.db.Exec(`
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
	err = s.db.QueryRow(`
        SELECT version_data 
        FROM wildberries.cards_history 
        WHERE nm_id = $1 AND version_id = $2
    `, nmID, versionID).Scan(&versionData)
	if err != nil {
		return err
	}

	// Шаг 3: Обновляем основную запись в cards_actual данными из версии
	_, err = s.db.Exec(`
        UPDATE wildberries.cards_actual
        SET version = version + 1, version_data = $1, updated_at = NOW()
        WHERE nm_id = $2
    `, versionData, nmID)
	return err
}

func (s *UpdateService) WaitForUpdateSignal(nmID int, version int, description string) {
	go func() {
		for {
			select {
			case <-s.updateCh:
				_, err := s.db.Exec(`
                    INSERT INTO wildberries.changes (nm_id, version, changed_at, description)
                    VALUES ($1, $2, NOW(), $3)
                `, nmID, version, description)
				if err != nil {
					log.Println("Error logging update:", err)
				} else {
					log.Println("Update logged successfully")
				}
			}
		}
	}()
}

// Параллельный сервис, который опрашивает API и отправляет сигнал
func checkForUpdates(updateCh chan bool) {
	for {
		// Симуляция проверки обновления через API
		time.Sleep(5 * time.Second)
		updateCh <- true
	}
}
