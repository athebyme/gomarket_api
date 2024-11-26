package dbconnect

import "database/sql"

type Database interface {
	Connect() (*sql.DB, error)
	Ping() error
}
