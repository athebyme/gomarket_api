package handlers

import "database/sql"

// прописать хендлеры для каждого эндпоинта !

type Handler interface {
	Connect() (*sql.DB, error)
	Ping() error
}
