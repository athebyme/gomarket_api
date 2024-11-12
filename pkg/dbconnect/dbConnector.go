package dbconnect

import "database/sql"

type DbConnector interface {
	Connect() (*sql.DB, error)
}
