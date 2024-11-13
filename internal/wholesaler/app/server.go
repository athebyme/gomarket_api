package app

import (
	"database/sql"
	"encoding/json"
	"gomarketplace_api/internal/wholesaler/internal/business"
	"gomarketplace_api/internal/wholesaler/internal/models/requests"
	"gomarketplace_api/internal/wholesaler/internal/storage"
	"gomarketplace_api/pkg/dbconnect"
	"gomarketplace_api/pkg/dbconnect/migration"
	"log"
	"net/http"
	"time"
)

type WholesalerServer struct {
	dbconnect.DbConnector
}

func NewWServer(dbCon dbconnect.DbConnector) *WholesalerServer {
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

	mediaRepo := storage.NewMediaRepository(productRepo)
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

// прописать хендлеры для каждого эндпоинта !
type HandlerInterface interface {
	Connect() (*sql.DB, error)
	Ping() error
}

type ProductHandler struct {
	dbconnect.DbConnector
	productService *business.ProductService
}

func NewProductHandler(connector dbconnect.DbConnector) *ProductHandler {
	db, err := connector.Connect()
	if err != nil {
		return nil
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

	return &ProductHandler{
		DbConnector:    connector,
		productService: productService,
	}
}

func (h *ProductHandler) Connect() (*sql.DB, error) {
	return h.DbConnector.Connect()
}

func (h *ProductHandler) Ping() error {
	err := h.DbConnector.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (h *ProductHandler) GetGlobalIDsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := h.Connect()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	if err = db.Ping(); err != nil {
		return
	}

	globalIDs, err := h.productService.GetAllGlobalIDs()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(globalIDs)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func (h *ProductHandler) GetAppellationHandler(w http.ResponseWriter, r *http.Request) {
	db, err := h.Connect()
	if err != nil {
		return
	}

	if err = db.Ping(); err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	appellations, err := h.productService.GetAllAppellations()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(appellations)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

func (h *ProductHandler) GetDescriptionsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := h.Connect()
	if err != nil {
		return
	}

	if err = db.Ping(); err != nil {
		return
	}

	descriptions, err := h.productService.GetAllDescriptions()
	if err != nil {
		http.Error(w, "Failed to fetch Global IDs", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(descriptions)
	if err != nil {
		log.Printf("Failed to fetch Appellations: %v", err)
	}
}

type MediaHandler struct {
	dbconnect.DbConnector
	repo *storage.MediaRepository
}

func NewMediaHandler(connector dbconnect.DbConnector) *MediaHandler {
	db, err := connector.Connect()
	if err != nil {
		return nil
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
	return &MediaHandler{
		DbConnector: connector,
		repo:        storage.NewMediaRepository(productRepo),
	}
}

func (h *MediaHandler) GetMediaHandler(w http.ResponseWriter, r *http.Request) {

	if err := h.Ping(); err != nil {
		http.Error(w, "Failed to ping database", http.StatusInternalServerError)
		return
	}

	// Декодирование тела запроса
	var mediaReq requests.MediaRequest
	if err := json.NewDecoder(r.Body).Decode(&mediaReq); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var mediaMap map[int][]string
	var err error

	startTime := time.Now()
	if len(mediaReq.ProductIDs) == 0 { // Проверка наличия productIDs
		mediaMap, err = h.repo.GetMediaSources(mediaReq.Censored)
		if err != nil {
			http.Error(w, "Failed to fetch all media sources", http.StatusInternalServerError)
			return
		}
	} else {
		mediaMap, err = h.repo.GetMediaSourcesByProductIDs(mediaReq.ProductIDs, mediaReq.Censored)
		if err != nil {
			http.Error(w, "Failed to fetch media sources", http.StatusInternalServerError)
			return
		}
	}
	log.Printf("media source execution time: %v", time.Since(startTime))

	// Кодирование ответа
	err = json.NewEncoder(w).Encode(mediaMap)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func SetupRoutes(handlers ...HandlerInterface) {
	// Создаем карту для хранения обработчиков по их типам
	handlerMap := make(map[string]HandlerInterface)

	// Заполняем карту обработчиков
	for _, handler := range handlers {
		switch h := handler.(type) {
		case *ProductHandler:
			handlerMap["ProductHandler"] = h
		case *MediaHandler:
			handlerMap["MediaHandler"] = h
		default:
			log.Printf("Unknown handler type: %T", h)
		}
	}

	// Проверяем наличие необходимых обработчиков и вызываем Ping для каждого
	for _, handler := range handlerMap {
		if err := handler.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
	}

	// Проверка и настройка маршрутов для ProductHandler
	if productHandler, ok := handlerMap["ProductHandler"].(*ProductHandler); ok {
		http.HandleFunc("/api/globalids", productHandler.GetGlobalIDsHandler)
		http.HandleFunc("/api/appellations", productHandler.GetAppellationHandler)
		http.HandleFunc("/api/descriptions", productHandler.GetDescriptionsHandler)
	} else {
		log.Fatalf("ProductHandler not provided")
	}

	if mediaHandler, ok := handlerMap["MediaHandler"].(*MediaHandler); ok {
		http.HandleFunc("/api/media", mediaHandler.GetMediaHandler)
	} else {
		log.Fatalf("MediaHandler not provided")
	}

	log.Printf("Запущен сервис wholesaler /api/")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
