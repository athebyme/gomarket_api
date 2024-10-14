package main

import (
	"gomarketplace_api/build/_postgres"
	"gomarketplace_api/internal/wholesaler/products"
	"gomarketplace_api/pkg/dbconnect/migration"
	"gomarketplace_api/pkg/dbconnect/postgres"
	"log"
)

func main() {
	log.Printf("\nStarted App\n")

	var db, err = postgres.ConnectToPostgreSQL()
	if err != nil {
		log.Printf("Error connecting to PostgreSQL: %s\n", err)
	}
	defer db.Close()

	migrationApply := []migration.MigrationInterface{
		&products.WholesalerSchema{},
		&products.WholesalerProducts{},
		&products.WholesalerDescriptions{},
		&products.WholesalerPrice{},
		&products.WholesalerStock{},

		&_postgres.CreateWildberriesSchema{},
		&_postgres.CreateCategoriesTable{},
		&_postgres.CreateProductsTable{},
	}

	for _, _migration := range migrationApply {
		if err := _migration.UpMigration(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("Migrations applied successfully!")
}
