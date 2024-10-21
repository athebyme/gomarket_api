package app

import (
	business2 "gomarketplace_api/internal/wholesaler/internal/business"
	"gomarketplace_api/internal/wholesaler/internal/products"
	storage2 "gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/pkg/dbconnect/migration"
	"gomarketplace_api/pkg/dbconnect/postgres"
	"log"
)

type WholesalerServer struct {
}

func NewWServer() *WholesalerServer {
	return &WholesalerServer{}
}

func (s *WholesalerServer) Run() {
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
	}

	for _, _migration := range migrationApply {
		if err := _migration.UpMigration(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("Wholesaler migrations applied successfully!")

	productRepo, err := storage2.NewProductRepository()
	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
	}
	productService := business2.NewProductService(productRepo)
	defer productRepo.Close()

	priceRepo, err := storage2.NewPriceRepository()
	if err != nil {
		log.Fatalf("Failed to create product repository: %v", err)
	}
	priceService := business2.NewPriceService(priceRepo)
	defer priceRepo.Close()

	stocksRepo, err := storage2.NewStocksRepository()
	if err != nil {
		log.Fatalf("Failed to create stocks repository: %v", err)
	}
	stocksService := business2.NewStockService(stocksRepo)
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
}
