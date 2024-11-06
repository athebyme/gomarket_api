package storage

import (
	"database/sql"
	"log"
)

type DescriptionsRepository struct {
	db      *sql.DB
	updater Updater
}

func NewDescriptionsRepository(db *sql.DB, updater Updater) *DescriptionsRepository {
	log.Println("Successfully connected to wholesaler descriptions repository")
	return &DescriptionsRepository{db: db, updater: updater}
}

func (r *DescriptionsRepository) Update(args ...[]string) error {
	return r.updater.Update(args...)
}
