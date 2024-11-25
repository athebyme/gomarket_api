package repositories

import (
	"database/sql"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"log"
)

type DescriptionsRepository struct {
	db      *sql.DB
	updater storage.Updater
}

func NewDescriptionsRepository(db *sql.DB, updater storage.Updater) *DescriptionsRepository {
	log.Println("Successfully connected to wholesaler descriptions repository")
	return &DescriptionsRepository{db: db, updater: updater}
}

func (r *DescriptionsRepository) Update(args ...[]string) error {
	return r.updater.Update(args...)
}

func (r *DescriptionsRepository) Close() error {
	return r.db.Close()
}
