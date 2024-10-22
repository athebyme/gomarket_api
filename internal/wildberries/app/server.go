package app

import (
	"gomarketplace_api/build/_postgres"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wildberries/internal/business/dto/responses"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	"gomarketplace_api/internal/wildberries/internal/business/services/update"
	"gomarketplace_api/pkg/dbconnect/migration"
	"gomarketplace_api/pkg/dbconnect/postgres"
	"log"
)

type WildberriesServer struct {
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

	log.Printf("\n\n\n\nGetting cards")

	var nomenclatures *responses.NomenclatureResponse
	nomenclatures, err = update.GetNomenclature(request.Settings{
		Sort:   request.Sort{Ascending: true},
		Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 100},
	}, "")
	if err != nil {
		log.Fatalf("Error getting nomenclatures: %s\n", err)
	}
	for _, v := range nomenclatures.Data {
		log.Printf("\nItem: %s", v.NmUUID)
		log.Printf("\nNomenclature : %v", v.NmUUID)
		log.Printf("\nPhotos: %v", v.Photos)
		log.Println()
	}
}

func checkFunctionality() {
	var err error
	var ping *responses.Ping
	ping, err = get.Ping()
	if err != nil {
		log.Fatalf("Error pingig WB server : %s\n", err)
	}
	log.Printf("WB server status is (%s), (TS: %s)", ping.Status, ping.TS)

	// ----
	cat_id := 5067
	var charcs *responses.CharacteristicsResponse
	charcs, err = get.GetItemCharcs(cat_id, "")
	if err != nil {
		log.Fatalf("Error getting characters : %s\n", err)
	}

	filtered := charcs.FilterPopularity()
	for _, v := range filtered.Data {
		log.Printf("\nPopular characters : %s", v.Name)
	}

	var colors *responses.ColorResponse
	colors, err = get.GetColors("")
	if err != nil {
		log.Fatalf("Error getting colors: %s\n", err)
	}
	for _, v := range colors.Data {
		log.Printf("\nColor : %s", v.Name)
	}

	var sex *responses.SexResponse
	sex, err = get.GetSex("")
	if err != nil {
		log.Fatalf("Error getting sexes: %s\n", err)
	}
	for _, v := range sex.Data {
		log.Printf("\nSex : %s", v)
	}

	var countries *responses.CountryResponse
	countries, err = get.GetCountries("")
	if err != nil {
		log.Fatalf("Error getting countries: %s\n", err)
	}
	for _, v := range countries.Data {
		log.Printf("\nSex : %s", v)
	}

	var prodCardsLim *responses.ProductCardsLimitResponse
	prodCardsLim, err = get.GetProductCardsLimit()
	if err != nil {
		log.Fatalf("Error getting product cards limit: %s\n", err)
	}
	log.Printf("\nProduct cards limit : %v", prodCardsLim.Data)

}
