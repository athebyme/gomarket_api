package app

import (
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"gomarketplace_api/pkg/dbconnect"
	"gomarketplace_api/pkg/dbconnect/migration"
	"log"
)

type WholesalerServer struct {
	dbconnect.Database
}

func NewWServer(dbCon dbconnect.Database) *WholesalerServer {
	return &WholesalerServer{dbCon}
}

func (s *WholesalerServer) Run(wg *chan struct{}) {
	var db, err = s.Connect()
	if err != nil {
		log.Printf("Error connecting to PostgreSQL: %s\n", err)
	}
	defer db.Close()

	migrationApply := []migration.MigrationInterface{
		&storage.WholesalerSchema{},
		&storage.MigrationsSchema{},
		&storage.WholesalerProducts{},
		&storage.WholesalerDescriptions{},
		&storage.WholesalerPrice{},
		&storage.WholesalerStock{},
		&storage.WholesalerMedia{},
		&storage.ProductSize{},
		&storage.Metadata{},
	}

	for _, _migration := range migrationApply {
		if err := _migration.UpMigration(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("Wholesaler migrations applied successfully!")

	productSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_info.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_info.csv",
		LastUpdateColumn: "last_update_products"}

	productUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"products",
		[]string{"global_id", "model", "appellation", "category",
			"brand", "country", "product_type", "features",
			"sex", "color", "dimension", "package", "empty",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := repositories.NewProductRepository(db, productUpdater)
	err = productRepo.Update()
	if err != nil {
		log.Fatalf("Error updating products: %s\n", err)
	}

	mediaRepo := repositories.NewMediaRepository(productRepo)
	err = mediaRepo.PopulateMediaTable()
	if err != nil {
		log.Fatalf("Error populating media table: %s\n", err)
	}

	defer productRepo.Close()

	priceSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_prices.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_prices.csv",
		LastUpdateColumn: "last_update_prices"}

	priceUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"price",
		[]string{"id товара", "цена"},
		priceSource.LastUpdateColumn,
		priceSource.InfURL,
		priceSource.CSVURL)

	priceRepo := repositories.NewPriceRepository(db, priceUpdater)
	err = priceRepo.Update([]string{"global_id", "price"})
	if err != nil {
		return
	}
	defer priceRepo.Close()

	stockSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_prices__.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_prices__.csv",
		LastUpdateColumn: "last_update_stocks"}

	stockUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"stocks",
		[]string{"id товара", "наличие"},
		stockSource.LastUpdateColumn,
		stockSource.InfURL,
		stockSource.CSVURL)

	stocksRepo := repositories.NewStocksRepository(db, stockUpdater)
	err = stocksRepo.Update([]string{"global_id", "stocks"})
	if err != nil {
		return
	}
	defer stocksRepo.Close()

	descriptionSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_info.inf",
		CSVURL:           "http://www.sexoptovik.ru/files/all_prod_d33_.csv",
		LastUpdateColumn: "last_update_description"}

	descriptionRepo := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"descriptions",
		[]string{"global_id", "product_description"},
		descriptionSource.LastUpdateColumn,
		descriptionSource.InfURL,
		descriptionSource.CSVURL)
	err = descriptionRepo.Update()
	if err != nil {
		return
	}
	defer stocksRepo.Close()
	*wg <- struct{}{}
}
