package _postgres

import (
	"database/sql"
	"fmt"
	"log"
)

type CreateWBProductsTable struct{}

func (m *CreateWBProductsTable) UpMigration(db *sql.DB) error {
	if ok, err := checkAndSkipMigration(db, "wildberries.products"); err != nil {
		return err
	} else if ok {
		return nil
	}
	query := `
	CREATE TABLE IF NOT EXISTS wildberries.products (
		global_id INT PRIMARY KEY,
		category_id INT NOT NULL,
		distance FLOAT NOT NULL,
	    FOREIGN KEY (category_id) REFERENCES wildberries.categories(category_id)
	);`
	if err := executeAndMarkMigration(db, query, "wildberries.products"); err != nil {
		return err
	}
	log.Println("Migration 'wildberries.products' completed successfully.")
	return nil
}

type CreateWBCategoriesTable struct{}

func (m *CreateWBCategoriesTable) UpMigration(db *sql.DB) error {
	if ok, err := checkAndSkipMigration(db, "wildberries.categories"); err != nil {
		return err
	} else if ok {
		return nil
	}
	query := `
	CREATE TABLE IF NOT EXISTS wildberries.categories (
		category_id INT PRIMARY KEY,
		category VARCHAR(255) NOT NULL,
		parent_category_id INT,
		parent_category_name VARCHAR(255) NOT NULL
	);`
	if err := executeAndMarkMigration(db, query, "wildberries.categories"); err != nil {
		return err
	}
	log.Println("Migration 'wildberries.categories' completed successfully.")
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

/*
Вопрос: нужно ли это хранить в бд ?
*/
type WBNomenclatures struct{}

func (m *WBNomenclatures) UpMigration(db *sql.DB) error {
	if ok, err := checkAndSkipMigration(db, "wildberries.nomenclatures"); err != nil {
		return err
	} else if ok {
		return nil
	}

	query := `
		CREATE TABLE IF NOT EXISTS wildberries.nomenclatures (
		    nomenclature_id SERIAL PRIMARY KEY,
			global_id INT,
			nm_id INT UNIQUE,	
			imt_id INT UNIQUE,
            nm_uuid UUID UNIQUE,
            vendor_code VARCHAR(255),
			subject_id INT,
		    wb_brand VARCHAR(255),
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
	if ok, err := checkAndSkipMigration(db, "wildberries.characteristics"); err != nil {
		return err
	} else if ok {
		return nil
	}

	query := `
		CREATE TABLE IF NOT EXISTS wildberries.characteristics (
			id SERIAL PRIMARY KEY,
			charc_id INT UNIQUE,
		    name TEXT,
			required BOOLEAN,
			subject_id INT,
		    unit_name VARCHAR(15),
		    max_count INT,
			popular BOOLEAN,
		    charc_type INT,
			FOREIGN KEY (subject_id) REFERENCES wildberries.categories(category_id)
		);
	`

	if err := executeAndMarkMigration(db, query, "wildberries.characteristics"); err != nil {
		return err
	}
	log.Println("Migration 'wildberries.characteristics' completed successfully.")
	return nil
}

func checkAndSkipMigration(db *sql.DB, migrationName string) (bool, error) {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = $1)", migrationName).Scan(&migrationExists)
	if err != nil {
		return migrationExists, fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Printf("Migration '%s' already completed. Skipping.\n", migrationName)
		return migrationExists, nil
	}
	return migrationExists, nil
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
