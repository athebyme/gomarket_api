package app

import (
	"database/sql"
	"gomarketplace_api/build/_postgres"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	"gomarketplace_api/internal/wildberries/internal/business/services/update"
	"gomarketplace_api/internal/wildberries/internal/storage"
	"gomarketplace_api/pkg/business/service"
	"gomarketplace_api/pkg/dbconnect"
	"gomarketplace_api/pkg/dbconnect/migration"
	"gomarketplace_api/pkg/logger"
	"io"
	"log"
)

type WildberriesServer struct {
	nomenclatureService *update.CardUpdateService
	dbconnect.Database
	config.WildberriesConfig
	log    logger.Logger
	writer io.Writer
}

func NewWbServer(connector dbconnect.Database, wbConfig config.WildberriesConfig, writer io.Writer) *WildberriesServer {
	_log := logger.NewLogger(writer, "[WildberriesServer]")
	return &WildberriesServer{Database: connector, WildberriesConfig: wbConfig, log: _log, writer: writer}
}

func (s *WildberriesServer) Run(wg *chan struct{}) {
	<-*wg

	var authEngine services.AuthEngine
	authEngine = services.NewBearerAuth(s.ApiKey)

	var db, err = s.Connect()
	if err != nil {
		s.log.Log("Error connecting to PostgreSQL: %s\n", err)
	}
	defer db.Close()

	nomenclatureUpdGet := get.NewNomenclatureEngine(db, authEngine, s.writer)
	s.nomenclatureService = update.NewCardUpdateService(
		nomenclatureUpdGet,
		service.NewTextService(),
		"http://localhost:8081",
		authEngine,
		s.log,
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
	s.log.Log("WB migrations applied successfully!")

	//_, err = s.nomenclatureService.UpdateDBNomenclatures(request.Settings{
	//	Sort:   request.Sort{Ascending: false},
	//	Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
	//	Cursor: request.Cursor{Limit: 15000},
	//}, "")
	//if err != nil {
	//	return
	//}

	//err = s.loadCharcs(db, authEngine)
	//if err != nil {
	//	s.log.FatalLog("Error loading Charcs: %w\n", err)
	//}

	s.uploadProducts(authEngine)
}

func (s *WildberriesServer) updateNames() interface{} {
	s.log.Log("Naming updater ")
	updateMedia, err := s.nomenclatureService.UpdateCardNaming(request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: 1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 10000},
	})

	if err != nil {
		s.log.FatalLog("Error updating nomenclatures: %s\n", err)
	}
	s.log.Log("\n\n\nUpdated %d nomenclatures\n", updateMedia)

	return updateMedia
}

func (s *WildberriesServer) uploadProducts(auth services.AuthEngine) interface{} {
	wsUrl := "http://localhost:8081"
	textService := service.NewTextService()
	db, err := s.Database.Connect()
	if err != nil {
		s.log.FatalLog("Error connecting to PostgreSQL: %s\n", err)
	}

	engine := get.NewNomenclatureEngine(db, auth, s.writer)
	repo := storage.NewNomenclatureRepository(db)

	nmService := update.NewNomenclatureService(*engine, *repo)
	cardService := update.NewCardService(wsUrl, textService, s.writer, s.WildberriesConfig, *nmService)

	accuracy := float32(0.3)
	categoryID := 0
	result, err := nmService.GetSetOfUncreatedItemsWithCategories(accuracy, true, categoryID)

	ids := make([]int, len(result))
	for k, _ := range result {
		ids = append(ids, k)
	}

	resultIDs, err := cardService.PrepareAndUpload(ids)
	if err != nil {
		return nil
	}

	s.log.Log("Total count of items : %d", len(resultIDs.([]int)))
	return struct{}{}
}

func (s *WildberriesServer) loadCharcs(db *sql.DB, engine services.AuthEngine) error {
	charcUpdate := get.NewUpdateDBCharcs(db, *get.NewCharacteristicService(engine))
	catsRepo := get.NewDBCategories(db)
	cats, err := catsRepo.Categories()
	if err != nil {
		s.log.FatalLog(err.Error())
		return err
	}
	catIDs := make([]int, len(cats))
	for i, v := range cats {
		catIDs[i] = v.SubjectID
	}
	_, err = charcUpdate.UpdateDBCharcs(catIDs)
	if err != nil {
		s.log.FatalLog(err.Error())
		return err
	}
	return nil
}
