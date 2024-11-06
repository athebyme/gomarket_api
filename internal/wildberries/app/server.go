package app

import (
	"gomarketplace_api/build/_postgres"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	get2 "gomarketplace_api/internal/wildberries/internal/business/models/get"
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	"gomarketplace_api/internal/wildberries/internal/business/services/update"
	"gomarketplace_api/pkg/business/service"
	"gomarketplace_api/pkg/dbconnect/migration"
	"gomarketplace_api/pkg/dbconnect/postgres"
	"log"
)

type WildberriesServer struct {
	cardService *update.CardUpdater
}

func NewWbServer() *WildberriesServer {
	return &WildberriesServer{}
}

func (s *WildberriesServer) Run(wg *chan struct{}) {
	<-*wg
	cfg := config.GetMarketplaceConfig()
	if cfg.ApiKey == "" {
		log.Printf("wb api key not set\n")
	}
	var db, err = postgres.ConnectToPostgreSQL()
	if err != nil {
		log.Printf("Error connecting to PostgreSQL: %s\n", err)
	}
	defer db.Close()

	loader := service.NewPostgresLoader(db)
	ch := make(chan bool)
	updater := get2.NewUpdater(loader, ch)
	nomenclatureUpdGet := get.NewNomenclatureUpdateGetter(db, *updater)
	s.cardService = update.NewCardUpdater(
		nomenclatureUpdGet,
		service.NewTextService(),
		"http://localhost:8081",
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

	_, err = s.cardService.NomenclatureService.UploadToDb(request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 10500},
	}, "")
	if err != nil {
		log.Fatalf("Error getting Nomenclature count: %v", err)
	}

	res, err := s.cardService.NomenclatureService.GetNomenclaturesWithLimitConcurrency(108, "")
	if err != nil {
		log.Fatalf("Error getting Nomenclature count: %v", err)
	}

	log.Printf("nomenclatures : %d", len(res))

	//updateAppellations, err := s.cardService.UpdateCardNaming(request.Settings{
	//	Sort:   request.Sort{Ascending: false},
	//	Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
	//	Cursor: request.Cursor{Limit: 1500},
	//})
	//
	//if err != nil {
	//	log.Fatalf("Error updating nomenclatures: %s\n", err)
	//}
	//log.Printf("\n\n\nUpdated %d nomenclatures\n", updateAppellations)
}
