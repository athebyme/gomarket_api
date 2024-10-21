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

type CreateCategoriesTable struct{}

func (m *CreateCategoriesTable) UpMigration(db *sql.DB) error {
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
