package migration

import "database/sql"

type MigrationInterface interface {
	UpMigration(*sql.DB) error
}
