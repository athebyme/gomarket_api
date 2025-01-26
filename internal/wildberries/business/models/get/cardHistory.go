package get

import (
	"gomarketplace_api/pkg/business/service"
	"time"
)

type UpdateService struct {
	loader   service.DatabaseLoader
	updateCh chan bool // Канал для получения сигнала об изменении
}

func NewUpdater(loader service.DatabaseLoader, updateCh chan bool) *UpdateService {
	return &UpdateService{loader: loader, updateCh: updateCh}
}

func (s *UpdateService) InitializeCard(cardData map[string]interface{}) error {
	cardData["version"] = 1
	return s.loader.InsertCardActual(cardData)
}

// пока не доделано
func (s *UpdateService) WaitForUpdateSignal(nmID int, version int, description string) {
	go func() {
		for {
			select {
			case <-s.updateCh:
				//_, err := s.loader.Exec(`
				//    INSERT INTO wildberries.changes (nm_id, version, changed_at, description)
				//    VALUES ($1, $2, NOW(), $3)
				//`, nmID, version, description)
				//if err != nil {
				//	log.Println("Error logging update:", err)
				//} else {
				//	log.Println("Update logged successfully")
				//}
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
