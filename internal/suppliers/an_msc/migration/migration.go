package migration

import (
	"database/sql"
	"fmt"
	"log"
)

type AnMscSchemaMigration struct{}

func (m *AnMscSchemaMigration) UpMigration(db *sql.DB) error {
	query := `
	CREATE SCHEMA IF NOT EXISTS an_msc; 
	`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

type AnMscProductsMigration struct{}

func (m *AnMscProductsMigration) UpMigration(db *sql.DB) error {
	if ok, err := checkAndSkipMigration(db, "an_msc.products"); err != nil {
		return err
	} else if ok {
		return nil
	}

	query :=
		`
	CREATE TABLE an_msc.products (
    code VARCHAR(50) PRIMARY KEY,
    article VARCHAR(50),
    title TEXT NOT NULL,
    group_code VARCHAR(50),
    group_title VARCHAR(255),
    category_code INT,
    category_title TEXT,
    tmn DECIMAL(10,2),
    msk INT,
    nsk DECIMAL(10,2),
    start_price DECIMAL(10,2),
    price DECIMAL(10,2) NOT NULL,
    discount VARCHAR(20),
    image VARCHAR(255),
    image1 VARCHAR(255),
    image2 VARCHAR(255),
    material VARCHAR(100),
    size VARCHAR(50),
    length DECIMAL(10,2),
    width DECIMAL(10,2),
    color VARCHAR(255),
    weight DECIMAL(10,2),
    battery VARCHAR(50),
    waterproof INT,
    country VARCHAR(50),
    manufacturer VARCHAR(100),
    barcode VARCHAR(40) UNIQUE,
    new INT,
    hit INT,
    description TEXT,
    collection VARCHAR(255),
    video TEXT,
    url TEXT,
    rst VARCHAR(30),
    spb INT,
    fixed_price VARCHAR(30),
    pieces INT,
    brand_code VARCHAR(40),
    brand_title VARCHAR(255),
    created timestamp,
    three_d VARCHAR(30),
    width_packed DECIMAL(10,2),
    height_packed DECIMAL(10,2),
    length_packed DECIMAL(10,2),
    weight_packed DECIMAL(10,3),
    modification_code VARCHAR(50),
    images TEXT,
    retail_price DECIMAL(10,2),
    kdr VARCHAR(30),
    category_new_code VARCHAR(30),
    category_new_title TEXT,
    embed3d VARCHAR(255),
    minsk VARCHAR(30),
    ast VARCHAR(30),
    barcodes VARCHAR(255),
    retail_price_minsk DECIMAL(10,2),
	marked INT);`

	if err := executeAndMarkMigration(db, query, "an_msc.products"); err != nil {
		return err
	}
	log.Println("Migration 'an_msc.products' completed successfully.")
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
