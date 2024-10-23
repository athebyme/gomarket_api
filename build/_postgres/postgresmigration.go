package _postgres

import (
	"database/sql"
	"fmt"
	"log"
)

type CreateWBProductsTable struct{}

func (m *CreateWBProductsTable) UpMigration(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS wildberries.products (
		global_id INT PRIMARY KEY,
		category_id INT NOT NULL,
		distance FLOAT NOT NULL,
	    FOREIGN KEY (category_id) REFERENCES wildberries.categories(category_id)
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wildberries.products table: %w", err)
	}
	return nil
}

type CreateWBCategoriesTable struct{}

func (m *CreateWBCategoriesTable) UpMigration(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS wildberries.categories (
		category_id INT PRIMARY KEY,
		category VARCHAR(255) NOT NULL,
		parent_category_id INT,
		parent_category_name VARCHAR(255) NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wildberries.categories table: %w", err)
	}
	return nil
}

type CreateWBSchema struct{}

func (m *CreateWBSchema) UpMigration(db *sql.DB) error {
	query := `
	CREATE SCHEMA IF NOT EXISTS wildberries;`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema wildberries: %w", err)
	}
	return nil
}

type WBNomenclatures struct{}

func (m *WBNomenclatures) UpMigration(db *sql.DB) error {
	if err := checkAndSkipMigration(db, "wholesaler.products"); err != nil {
		return err
	}

	query := `
		CREATE TABLE IF NOT EXISTS wildberries.nomenclatures (
		    nomenclature_id SERIAL PRIMARY KEY,
			global_id INT,
			nm_id INT UNIQUE,т8ш
			imt_id INT UNIQUE,
            nm_uuid UUID UNIQUE,
            vendor_code VARCHAR(255),
			subject_id INT,
		    wb_brand VARCHAR(255),
		    package_length DOUBLE,
		    package_width DOUBLE,
		    package_height DOUBLE,
			created_at TIMESTAMP WITH TIME ZONE,
			updated_at TIMESTAMP WITH TIME ZONE,
			FOREIGN KEY(global_id) REFERENCES wholesaler.products(global_id)
		);
	`
	if err := executeAndMarkMigration(db, query, "wildberries.nomenclatures"); err != nil {
		return err
	}
	log.Println("Migration 'wildberries.nomenclatures' completed successfully.")
	return nil
}

type WholesalerCharacteristics struct{}

func (m *WholesalerCharacteristics) UpMigration(db *sql.DB) error {
	if err := checkAndSkipMigration(db, "wildberries.characteristics"); err != nil {
		return err
	}

	query := `
		CREATE TABLE IF NOT EXISTS wholesaler.characteristics (
			id SERIAL PRIMARY KEY,
			charc_id INT UNIQUE,
			name TEXT,
			required BOOLEAN,
			subject_id INT, -- Внешний ключ к категории
			popular BOOLEAN,
            -- Другие поля характеристики
			FOREIGN KEY (subject_id) REFERENCES wholesaler.categories(category_id)
		);
	`

	if err := executeAndMarkMigration(db, query, "wildberries.characteristics"); err != nil {
		return err
	}
	log.Println("Migration 'wildberries.characteristics' completed successfully.")
	return nil
}

func checkAndSkipMigration(db *sql.DB, migrationName string) error {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = $1)", migrationName).Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Printf("Migration '%s' already completed. Skipping.\n", migrationName)
		return nil
	}
	return nil
}

func executeAndMarkMigration(db *sql.DB, query string, migrationName string) error {
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute migration '%s': %w", migrationName, err)
	}
	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ($1, current_timestamp)", migrationName)
	if err != nil {
		return fmt.Errorf("failed to mark migration '%s' as complete: %w", migrationName, err)
	}
	return nil
}
