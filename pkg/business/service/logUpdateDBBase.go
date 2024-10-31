package service

type DatabaseLoader interface {
	// == инициализация, только версию ставим 1 руками
	InsertCardActual(cardData map[string]interface{}) error
	InsertCardHistory(cardData map[string]interface{}, version int) error
	SaveVersion(cardData map[string]interface{}) error
	RollbackCard(nmID int, versionID int) error
}
