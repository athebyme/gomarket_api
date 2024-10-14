package products

import (
	"database/sql"
	"fmt"
)

type WholesalerProducts struct{}

func (m *WholesalerProducts) UpMigration(db *sql.DB) error {
	query :=
		`
		CREATE TABLE IF NOT EXISTS wholesaler.products (
		globalID INT PRIMARY KEY,
		model VARCHAR(255),
		appellation TEXT,
		category TEXT,
		brand VARCHAR(255),
		country VARCHAR(255),
		productType VARCHAR(255),
		features VARCHAR(255),
		sex VARCHAR(255),
		color VARCHAR(255),
		dimension TEXT,
		package TEXT,
		media VARCHAR(255),
		barcodes VARCHAR(255),
		meterial VARCHAR(255),
		packageBattery TEXT
		);
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.products table: %w", err)
	}
	return nil
}

type WholesalerDescriptions struct{}

func (m *WholesalerDescriptions) UpMigration(db *sql.DB) error {
	query :=
		`	
			CREATE TABLE IF NOT EXISTS wholesaler.descriptions (
			globalID INT,
			productDescription TEXT,
			productAppellation TEXT,
			FOREIGN KEY (globalID) REFERENCES wholesaler.products(globalID)
		);
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.descriptions table: %w", err)
	}
	return nil
}

type WholesalerStock struct{}

func (m *WholesalerStock) UpMigration(db *sql.DB) error {
	query :=
		`
		CREATE TABLE IF NOT EXISTS wholesaler.stocks (
		    globalID INT,
		    mainArticular VARCHAR(255) NOT NULL,
		    stocks INT,
		    FOREIGN KEY (globalID) REFERENCES wholesaler.products(globalID)
		);
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.stocks table: %w", err)
	}
	return nil
}

type WholesalerPrice struct{}

func (m *WholesalerPrice) UpMigration(db *sql.DB) error {
	query :=
		`
		CREATE TABLE IF NOT EXISTS wholesaler.price (
		    globalID INT,
		    price INT,
		    FOREIGN KEY (globalID) REFERENCES wholesaler.products(globalID)
		);
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.price table: %w", err)
	}
	return nil
}

type WholesalerSchema struct{}

func (m *WholesalerSchema) UpMigration(db *sql.DB) error {
	query :=
		`
		CREATE SCHEMA IF NOT EXISTS wholesaler;
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.schema table: %w", err)
	}
	return nil
}
