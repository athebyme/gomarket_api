package _postgres

import (
	"database/sql"
	"fmt"
)

type CreateProductsTable struct{}

func (m *CreateProductsTable) UpMigration(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS wildberries.products (
		global_id INT PRIMARY KEY,
		appellation TEXT NOT NULL,
		categoryID INT NOT NULL,
		distance FLOAT NOT NULL,
	    FOREIGN KEY (categoryID) REFERENCES wildberries.categories(categoryID)
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wildberries.products table: %w", err)
	}
	return nil
}

type CreateCategoriesTable struct{}

func (m *CreateCategoriesTable) UpMigration(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS wildberries.categories (
	    categoryID INT PRIMARY KEY,
		category VARCHAR(255) NOT NULL,
		parentCategoryID INT,
		parentCategoryName VARCHAR(255) NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wildberries.categories table: %w", err)
	}
	return nil
}

type CreateWildberriesSchema struct{}

func (m *CreateWildberriesSchema) UpMigration(db *sql.DB) error {
	query := `
	CREATE SCHEMA IF NOT EXISTS wildberries;`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema wildberries: %w", err)
	}
	return nil
}
