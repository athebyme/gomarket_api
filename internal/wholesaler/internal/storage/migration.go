package storage

import (
	"database/sql"
	"fmt"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"log"
)

type WholesalerProducts struct{}

func (m *WholesalerProducts) UpMigration(db *sql.DB) error {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = 'wholesaler.products')").Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Println("Migration 'wholesaler.products' already completed. Skipping.")
		return nil
	}
	query :=
		`
		CREATE TABLE IF NOT EXISTS wholesaler.products (
		global_id INT PRIMARY KEY,
		model VARCHAR(255),
		appellation TEXT,
		category TEXT,
		brand VARCHAR(255),
		country VARCHAR(255),
		product_type VARCHAR(255),
		features VARCHAR(255),
		sex VARCHAR(255),
		color VARCHAR(255),
		dimension TEXT,
		package TEXT,
		empty VARCHAR(255),
		media VARCHAR(255),
		barcodes VARCHAR(255),
		material VARCHAR(255),
		package_battery TEXT
		);
		`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.products table: %w", err)
	}
	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ('wholesaler.products', current_timestamp)")
	if err != nil {
		return fmt.Errorf("failed to mark wholesaler.products migration as complete: %w", err)
	}

	log.Println("Migration 'wholesaler.products' completed successfully.")
	return nil
}

type WholesalerDescriptions struct{}

func (m *WholesalerDescriptions) UpMigration(db *sql.DB) error {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = 'wholesaler.descriptions')").Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Println("Migration 'wholesaler.descriptions' already completed. Skipping.")
		return nil
	}
	query :=
		`	
			CREATE TABLE IF NOT EXISTS wholesaler.descriptions (
			global_id INT UNIQUE,
			product_description TEXT,
			FOREIGN KEY (global_id) REFERENCES wholesaler.products(global_id)
		);
		`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.descriptions table: %w", err)
	}
	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ('wholesaler.descriptions', current_timestamp)")
	if err != nil {
		return fmt.Errorf("failed to mark wholesaler.descriptions migration as complete: %w", err)
	}

	log.Println("Migration 'wholesaler.descriptions' completed successfully.")
	return nil
}

type WholesalerMedia struct{}

func (m *WholesalerMedia) UpMigration(db *sql.DB) error {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = 'wholesaler.media')").Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Println("Migration 'wholesaler.descriptions' already completed. Skipping.")
		return nil
	}
	query :=
		`	
			CREATE TABLE IF NOT EXISTS wholesaler.media (
			global_id INT,
			images TEXT[],
			images_censored TEXT[],
			FOREIGN KEY (global_id) REFERENCES wholesaler.products(global_id)
		);
		`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.media table: %w", err)
	}
	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ('wholesaler.media', current_timestamp)")
	if err != nil {
		return fmt.Errorf("failed to mark wholesaler.media migration as complete: %w", err)
	}

	log.Println("Migration 'wholesaler.media' completed successfully.")
	return nil
}

type WholesalerStock struct{}

func (m *WholesalerStock) UpMigration(db *sql.DB) error {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = 'wholesaler.stocks')").Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Println("Migration 'wholesaler.stocks' already completed. Skipping.")
		return nil
	}
	query :=
		`
		CREATE TABLE IF NOT EXISTS wholesaler.stocks (
		    global_id INT UNIQUE,
		    stocks INT,
		    FOREIGN KEY (global_id) REFERENCES wholesaler.products(global_id)
		);
		`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.stocks table: %w", err)
	}
	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ('wholesaler.stocks', current_timestamp)")
	if err != nil {
		return fmt.Errorf("failed to mark wholesaler.stocks migration as complete: %w", err)
	}

	log.Println("Migration 'wholesaler.stocks' completed successfully.")
	return nil
}

type WholesalerPrice struct{}

func (m *WholesalerPrice) UpMigration(db *sql.DB) error {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = 'wholesaler.price')").Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Println("Migration 'wholesaler.price' already completed. Skipping.")
		return nil
	}
	query :=
		`
		CREATE TABLE IF NOT EXISTS wholesaler.price (
		    global_id INT UNIQUE,
		    price INT,
		    FOREIGN KEY (global_id) REFERENCES wholesaler.products(global_id)
		);
		`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create wholesaler.price table: %w", err)
	}
	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ('wholesaler.price', current_timestamp)")
	if err != nil {
		return fmt.Errorf("failed to mark wholesaler.price migration as complete: %w", err)
	}

	log.Println("Migration 'wholesaler.price' completed successfully.")
	return nil
}

type ProductSize struct{}

func (m *ProductSize) UpMigration(db *sql.DB) error {
	err := migrateAndParse(db)
	if err != nil {
		return err
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

type MigrationsSchema struct{}

func (m *MigrationsSchema) UpMigration(db *sql.DB) error {
	query :=
		`
		CREATE SCHEMA IF NOT EXISTS migrations;
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations.migrations (
            id SERIAL PRIMARY KEY,
            time TIMESTAMP NOT NULL,
            name VARCHAR(255) UNIQUE NOT NULL
        );
    `)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

type Metadata struct{}

// UpMigration - создает таблицу metadata, если она еще не существует.
func (m *Metadata) UpMigration(db *sql.DB) error {
	// Проверяем, была ли применена миграция.
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = 'metadata')").Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Println("Migration 'metadata' already completed. Skipping.")
		return nil
	}

	// Создаем таблицу metadata, если она еще не существует.
	query := `
		CREATE TABLE IF NOT EXISTS metadata (
		    id SERIAL PRIMARY KEY,
		    key_name VARCHAR(255) UNIQUE NOT NULL,
		    value TEXT,
		    last_update TIMESTAMP
		);
	`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create metadata table: %w", err)
	}

	// Добавляем запись о миграции в таблицу migrations.
	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ('metadata', current_timestamp)")
	if err != nil {
		return fmt.Errorf("failed to mark metadata migration as complete: %w", err)
	}

	log.Println("Migration 'metadata' completed successfully.")
	return nil
}

func migrateAndParse(db *sql.DB) error {
	var migrationExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations.migrations WHERE name = 'wholesaler.size')").Scan(&migrationExists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	if migrationExists {
		log.Println("Migration 'wholesaler.size' already completed. Skipping.")
		return nil
	}
	sizeTableQuery := `
		CREATE TABLE IF NOT EXISTS wholesaler.sizes (
			size_id SERIAL PRIMARY KEY,
			global_id INT,
			FOREIGN KEY (global_id) REFERENCES wholesaler.products(global_id)
		);
	`

	_, err = db.Exec(sizeTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create sizes table: %w", err)
	}

	sizeValuesTableQuery := `
		CREATE TABLE IF NOT EXISTS wholesaler.size_values (
			value_id SERIAL PRIMARY KEY,
			size_id INT,
			descriptor VARCHAR(255),
			value_type VARCHAR(10),  -- 'COMMON', 'MIN', 'MAX'
			value DECIMAL(10, 2),
		    unit VARCHAR(10),
			FOREIGN KEY (size_id) REFERENCES wholesaler.sizes(size_id)
		);
	`
	_, err = db.Exec(sizeValuesTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create size_values table: %w", err)
	}

	productsQuery := `SELECT global_id, dimension FROM wholesaler.products`
	rows, err := db.Query(productsQuery)
	if err != nil {
		return fmt.Errorf("failed to select from wholesaler.products: %w", err)
	}
	defer rows.Close()

	insertSizeStmt, err := db.Prepare(`
		INSERT INTO wholesaler.sizes (global_id)
		VALUES ($1) RETURNING size_id
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement for sizes: %w", err)
	}
	defer insertSizeStmt.Close()

	insertSizeValueStmt, err := db.Prepare(`
		INSERT INTO wholesaler.size_values (size_id, descriptor, value_type, value, unit)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement for size_values: %w", err)
	}
	defer insertSizeValueStmt.Close()
	err = processProducts(db)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO migrations.migrations (name, time) VALUES ('wholesaler.size', current_timestamp)")
	if err != nil {
		return fmt.Errorf("failed to mark migration as complete: %w", err)
	}

	log.Println("Migration 'sizes' completed successfully.")
	return nil
}

func processProducts(db *sql.DB) error {
	productsQuery := `SELECT global_id, dimension FROM wholesaler.products`
	rows, err := db.Query(productsQuery)
	if err != nil {
		return fmt.Errorf("failed to select from wholesaler.products: %w", err)
	}
	defer rows.Close() // Обязательно закрываем rows

	// Подготавливаем запросы *вне* цикла
	insertSizeStmt, err := db.Prepare(
		"INSERT INTO wholesaler.sizes (global_id) VALUES ($1) RETURNING size_id")
	if err != nil {
		return fmt.Errorf("failed to prepare insertSize statement: %w", err)
	}
	defer insertSizeStmt.Close()

	insertSizeValueStmt, err := db.Prepare("INSERT INTO wholesaler.size_values (size_id, descriptor, value_type, value, unit) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("failed to prepare insertSizeValue statement: %w", err)
	}
	defer insertSizeValueStmt.Close()

	for rows.Next() {
		var globalID int
		var dimension string
		if err := rows.Scan(&globalID, &dimension); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		sizeDescriptors, err := ParseSizes(dimension) // Передаем dimension как строку
		if err != nil {
			return fmt.Errorf("failed to parse sizes for globalID %d: %w", globalID, err) // Более информативное сообщение
		}

		sizeMap := make(map[models.SizeDescriptorEnum]*float64)
		for _, size := range sizeDescriptors {
			sizeMap[size.Descriptor] = &size.Value
		}

		var sizeID int
		err = insertSizeStmt.QueryRow(globalID).Scan(&sizeID)
		if err != nil {
			return fmt.Errorf("failed to insert into sizes for globalID %d: %w", globalID, err) // Более информативное сообщение
		}

		for _, size := range sizeDescriptors {
			if size.Unit == "" {
				log.Printf("UNIT WARN : %s for %d is not defined", size.Unit, sizeID)
			}
			_, err = insertSizeValueStmt.Exec(sizeID, size.Descriptor, size.Type, size.Value, size.Unit)
			if err != nil {
				return fmt.Errorf("failed to insert into size_values for globalID %d: %w", globalID, err) // Более информативное сообщение
			}
		}
	}

	return rows.Err() // Обрабатываем ошибки итерации по rows
}
