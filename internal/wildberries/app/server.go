package app

import (
	"gomarketplace_api/build/_postgres"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"gomarketplace_api/pkg/dbconnect/migration"
	"gomarketplace_api/pkg/dbconnect/postgres"
	"log"
)

type WildberriesServer struct {
}

func NewWbServer() *WildberriesServer {
	return &WildberriesServer{}
}

func (s *WildberriesServer) Run() {
	cfg := config.GetMarketplaceConfig()
	if cfg.ApiKey == "" {
		log.Printf("wb api key not set\n")
	}
	var db, err = postgres.ConnectToPostgreSQL()
	if err != nil {
		log.Printf("Error connecting to PostgreSQL: %s\n", err)
	}
	defer db.Close()

	migrationApply := []migration.MigrationInterface{
		&_postgres.CreateWildberriesSchema{},
		&_postgres.CreateCategoriesTable{},
		&_postgres.CreateProductsTable{},
	}

	for _, _migration := range migrationApply {
		if err := _migration.UpMigration(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("WB migrations applied successfully!")

	var ping *responses.Ping
	ping, err = services.Ping()
	if err != nil {
		log.Fatalf("Error pingig WB server : %s\n", err)
	}
	log.Printf("WB server status is (%s), (TS: %s)", ping.Status, ping.TS)

	// ----
	cat_id := 5067
	var charcs *responses.CharacteristicsResponse
	charcs, err = services.GetItemCharcs(cat_id, "")
	if err != nil {
		log.Fatalf("Error getting characters : %s\n", err)
	}

	filtered := charcs.FilterPopularity()
	for _, v := range filtered.Data {
		log.Printf("\nPopular characters : %s", v.Name)
	}
}
