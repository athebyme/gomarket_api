package app

import (
	"gomarketplace_api/build/_postgres"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	get2 "gomarketplace_api/internal/wildberries/internal/business/models/get"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	"gomarketplace_api/internal/wildberries/internal/business/services/update"
	"gomarketplace_api/pkg/business/service"
	"gomarketplace_api/pkg/dbconnect"
	"gomarketplace_api/pkg/dbconnect/migration"
	"log"
)

type WildberriesServer struct {
	cardService *update.CardUpdater
	dbconnect.DbConnector
	config.MarketplaceConfig
}

func NewWbServer(connector dbconnect.DbConnector, marketplaceConfig config.MarketplaceConfig) *WildberriesServer {
	return &WildberriesServer{DbConnector: connector, MarketplaceConfig: marketplaceConfig}
}

func (s *WildberriesServer) Run(wg *chan struct{}) {
	<-*wg

	var authEngine services.AuthEngine
	authEngine = services.NewBearerAuth(s.ApiKey())

	var db, err = s.Connect()
	if err != nil {
		log.Printf("Error connecting to PostgreSQL: %s\n", err)
	}
	defer db.Close()

	loader := service.NewPostgresLoader(db)
	ch := make(chan bool)
	updater := get2.NewUpdater(loader, ch)
	nomenclatureUpdGet := get.NewNomenclatureUpdateGetter(db, *updater, authEngine)
	s.cardService = update.NewCardUpdater(
		nomenclatureUpdGet,
		service.NewTextService(),
		"http://localhost:8081",
		authEngine,
	)

	migrationApply := []migration.MigrationInterface{
		&_postgres.CreateWBSchema{},
		&_postgres.CreateWBCategoriesTable{},
		&_postgres.CreateWBProductsTable{},
		&_postgres.WBCharacteristics{},
		&_postgres.WBNomenclatures{},
		&_postgres.WBCardsActual{},
		&_postgres.WBNomenclaturesHistory{},
		&_postgres.WBChanges{},
	}

	for _, _migration := range migrationApply {
		if err := _migration.UpMigration(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("WB migrations applied successfully!")

	log.SetPrefix("Naming updater ")
	updateMedia, err := s.cardService.UpdateCardMedia(request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: 1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 1},
	})
	log.SetPrefix("")

	if err != nil {
		log.Fatalf("Error updating nomenclatures: %s\n", err)
	}
	log.Printf("\n\n\nUpdated %d nomenclatures\n", updateMedia)
}
