package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wholesaler/internal/business"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/pkg/dbconnect/migration"
	"gomarketplace_api/pkg/dbconnect/postgres"
	"log"
	"net/http"
)

type WholesalerServer struct {
}

func NewWServer() *WholesalerServer {
	return &WholesalerServer{}
}

func (s *WholesalerServer) Run(wg *chan struct{}) {
	var db, err = postgres.ConnectToPostgreSQL()
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
			"sex", "color", "dimension", "package",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := storage.NewProductRepository(db, productUpdater)
	err = productRepo.Update()
	if err != nil {
		log.Fatalf("Error updating products: %s\n", err)
	}

	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
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

	priceRepo := storage.NewPriceRepository(db, priceUpdater)
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

	stocksRepo := storage.NewStocksRepository(db, stockUpdater)
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

func getGlobalIDsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return
	}

	if err = db.Ping(); err != nil {
		return
	}

	productSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_info.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_info.csv",
		LastUpdateColumn: "last_update_products"}

	productUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"product",
		[]string{"global_id", "model", "appellation", "category",
			"brand", "country", "product_type", "features",
			"sex", "color", "dimension", "package",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := storage.NewProductRepository(db, productUpdater)
	productService := business.NewProductService(productRepo)
	globalIDs, err := productService.GetAllGlobalIDs()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	// ?
	defer productRepo.Close()

	err = json.NewEncoder(w).Encode(globalIDs)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func getAppellationHandler(w http.ResponseWriter, r *http.Request) {

	cfg := config.GetConfig()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return
	}

	if err = db.Ping(); err != nil {
		return
	}

	productSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_info.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_info.csv",
		LastUpdateColumn: "last_update_products"}

	productUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"product",
		[]string{"global_id", "model", "appellation", "category",
			"brand", "country", "product_type", "features",
			"sex", "color", "dimension", "package",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := storage.NewProductRepository(db, productUpdater)

	productService := business.NewProductService(productRepo)
	appellations, err := productService.GetAllAppellations()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	// ?
	defer productRepo.Close()

	err = json.NewEncoder(w).Encode(appellations)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func getDescriptionsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return
	}

	if err = db.Ping(); err != nil {
		return
	}

	productSource := storage.DataSource{
		InfURL:           "http://sexoptovik.ru/files/all_prod_info.inf",
		CSVURL:           "http://sexoptovik.ru/files/all_prod_info.csv",
		LastUpdateColumn: "last_update_products"}

	productUpdater := storage.NewPostgresUpdater(
		db,
		"wholesaler",
		"product",
		[]string{"global_id", "model", "appellation", "category",
			"brand", "country", "product_type", "features",
			"sex", "color", "dimension", "package",
			"media", "barcodes", "material", "package_battery"},
		productSource.LastUpdateColumn,
		productSource.InfURL,
		productSource.CSVURL)

	productRepo := storage.NewProductRepository(db, productUpdater)

	productService := business.NewProductService(productRepo)
	descriptions, err := productService.GetAllDescriptions()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	// ?
	defer productRepo.Close()

	err = json.NewEncoder(w).Encode(descriptions)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func SetupRoutes() {
	http.HandleFunc("/api/globalids", getGlobalIDsHandler)
	http.HandleFunc("/api/appellations", getAppellationHandler)
	http.HandleFunc("/api/descriptions", getDescriptionsHandler)
	log.Printf("Запущен сервис /api/globalids")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
