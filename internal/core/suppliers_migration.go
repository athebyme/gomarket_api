package core

import (
	"database/sql"
	"fmt"
	"log"
)

type CoreSuppliers struct{}

func (m *CoreSuppliers) UpMigration(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM core.suppliers WHERE name = 'wholesaler')").Scan(&exists)

	if err != nil {
		return fmt.Errorf("supplier check failed: %w", err)
	}

	if exists {
		log.Println("Supplier 'wholesaler' already exists")
		return nil
	}

	query := `
    CREATE SCHEMA IF NOT EXISTS core;
    
    CREATE TABLE IF NOT EXISTS core.suppliers (
        supplier_id SERIAL PRIMARY KEY,
        name VARCHAR(255) UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW()
    );
    
    INSERT INTO core.suppliers (name) 
    VALUES ('wholesaler') 
    ON CONFLICT DO NOTHING;`

	_, err = db.Exec(query)
	return err
}

type CoreProducts struct{}

func (m *CoreProducts) UpMigration(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS core.products (
        global_id INT PRIMARY KEY,
        supplier_id INT NOT NULL REFERENCES core.suppliers(supplier_id),
        base_data JSONB,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW()
    );
    
    -- Переносим существующие продукты в core
    INSERT INTO core.products (global_id, supplier_id)
    SELECT p.global_id, s.supplier_id 
    FROM wholesaler.products p
    CROSS JOIN core.suppliers s
    WHERE s.name = 'wholesaler'
    ON CONFLICT DO NOTHING;`

	_, err := db.Exec(query)
	return err
}

type CoreRelations struct{}

func (m *CoreRelations) UpMigration(db *sql.DB) error {
	queries := []string{
		`ALTER TABLE wholesaler.products 
         ADD FOREIGN KEY (global_id) REFERENCES core.products(global_id);`,

		`ALTER TABLE wildberries.nomenclatures 
         ADD FOREIGN KEY (global_id) REFERENCES core.products(global_id);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
