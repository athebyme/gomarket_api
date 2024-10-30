package app

import (
	"encoding/json"
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
	}

	for _, _migration := range migrationApply {
		if err := _migration.UpMigration(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("Wholesaler migrations applied successfully!")

	productRepo, err := storage.NewProductRepository()
	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
	}
	productService := business.NewProductService(productRepo)
	defer productRepo.Close()

	priceRepo, err := storage.NewPriceRepository()
	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
	}
	priceService := business.NewPriceService(priceRepo)
	defer priceRepo.Close()

	stocksRepo, err := storage.NewStocksRepository()
	if err != nil {
		log.Fatalf("Failed to create stocks repository: %v", err)
	}
	stocksService := business.NewStockService(stocksRepo)
	defer stocksRepo.Close()

	id := []int{9575, 1, 9574, 9778}
	prods, err := productService.GetProductsByIDs(id)
	if err != nil {
		log.Fatalf("Failed to get product: %v", err)
	}
	for _, v := range prods {
		log.Printf("Retrieved product: %+v", v)
		price, err := priceService.GetProductPriceByID(v.ID)
		if err != nil {
			log.Fatalf("Failed to get product price: %v", err)
		}

		stocks, err := stocksService.GetProductStocksByID(v.ID)
		if err != nil {
			log.Fatalf("Failed to get product stocks: %v", err)
		}
		log.Printf(
			"Its price : %d. Its stocks : %d. Main-articular : %s",
			price.Price, stocks.Stocks, stocks.MainArticular,
		)
	}
	*wg <- struct{}{}
}

func getGlobalIDsHandler(w http.ResponseWriter, r *http.Request) {
	repo, err := storage.NewProductRepository()
	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
		return
	}
	productService := business.NewProductService(repo)
	globalIDs, err := productService.GetAllGlobalIDs()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	// ?
	defer repo.Close()

	err = json.NewEncoder(w).Encode(globalIDs)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func getAppellationHandler(w http.ResponseWriter, r *http.Request) {
	repo, err := storage.NewProductRepository()
	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
		return
	}
	productService := business.NewProductService(repo)
	appellations, err := productService.GetAllAppellations()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	// ?
	defer repo.Close()

	err = json.NewEncoder(w).Encode(appellations)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func getDescriptionsHandler(w http.ResponseWriter, r *http.Request) {
	repo, err := storage.NewProductRepository()
	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
		return
	}
	productService := business.NewProductService(repo)
	descriptions, err := productService.GetAllDescriptions()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	// ?
	defer repo.Close()

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
