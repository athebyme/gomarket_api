package app

import (
	"context"
	"database/sql"
	"gomarketplace_api/build/_postgres"
	"gomarketplace_api/config"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/request"
	"gomarketplace_api/internal/wildberries/internal/business/services"
	"gomarketplace_api/internal/wildberries/internal/business/services/get"
	"gomarketplace_api/internal/wildberries/internal/business/services/parse"
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
	cardUpdateService *update.CardUpdateService
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

	nomenclatureUpdGet := get.NewSearchEngine(db, authEngine, s.writer)
	s.cardUpdateService = update.NewCardUpdateService(
		nomenclatureUpdGet,
		service.NewTextService(),
		"http://localhost:8081",
		authEngine,
		s.log,
		parse.NewBrandServiceWildberries(s.WbBanned.BannedBrands),
		s.WbValues,
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

	//err = s.loadCharcs(db, authEngine)
	//if err != nil {
	//	s.log.FatalLog("Error loading Charcs: %w\n", err)
	//}

	//categories := []int{2865, 5071, 5073, 5067}
	//for _, categoryID := range categories {
	//	s.uploadProducts(authEngine, categoryID)
	//	time.Sleep(30 * time.Second)
	//}
	//

	//
	//time.Sleep(5 * time.Second)
	//
	_, err = s.cardUpdateService.UpdateDBNomenclatures(request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 10000},
	}, "")
	if err != nil {
		return
	}

	//_, err = s.cardUpdateService.UpdateCardMedia(request.Settings{
	//	Sort:   request.Sort{Ascending: false},
	//	Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
	//	Cursor: request.Cursor{Limit: 10500},
	//})
	//err = s.updateByCategoryId()

	//_, err = s.cardUpdateService.UpdateCardNaming(request.Settings{
	//	Sort:   request.Sort{Ascending: false},
	//	Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{}, Brands: []string{}, ImtID: 0},
	//	Cursor: request.Cursor{Limit: 10000},
	//})
	//if err != nil {
	//	return
	//}

}

func (s *WildberriesServer) updateNames() interface{} {
	s.log.Log("Naming updater ")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updateMedia, err := s.cardUpdateService.UpdateCardNaming(ctx, request.Settings{
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

func (s *WildberriesServer) uploadProducts(auth services.AuthEngine, categoryID int) interface{} {
	wsUrl := "http://localhost:8081"
	textService := service.NewTextService()
	db, err := s.Database.Connect()
	if err != nil {
		s.log.FatalLog("Error connecting to PostgreSQL: %s\n", err)
	}

	engine := get.NewSearchEngine(db, auth, s.writer)
	repo := storage.NewNomenclatureRepository(db)

	nmService := update.NewNomenclatureService(*engine, *repo)
	cardService := update.NewCardService(wsUrl, textService, s.writer, s.WildberriesConfig)

	accuracy := float32(0.3)
	result, err := nmService.GetSetOfUncreatedItemsWithCategories(accuracy, true, categoryID)

	ids := make([]int, len(result))
	for k, _ := range result {
		ids = append(ids, k)
	}

	resultIDs, err := cardService.PrepareAndUpload(ids)
	if err != nil {
		return nil
	}

	req := []request.CreateCardRequestWrapper{}
	for i, v := range resultIDs.([]request.CreateCardRequestData) {
		req = append(req, request.CreateCardRequestWrapper{})
		req[i].Variants = append(req[i].Variants, v)
		req[i].SubjectID = categoryID
	}

	_, _, err = cardService.SendToServerModels(req)
	if err != nil {
		return nil
	}

	s.log.Log("Category %d. Total count of items : %d", categoryID, len(resultIDs.([]request.CreateCardRequestData)))
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

func (s *WildberriesServer) updateByCategoryId() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updateAppellation, err := s.cardUpdateService.UpdateCardNaming(ctx, request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{5067}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 1500},
	})
	updatePackages, err := s.cardUpdateService.UpdateCardPackages(request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{5067}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 1500},
	})
	updateMedia, err := s.cardUpdateService.UpdateCardMedia(request.Settings{
		Sort:   request.Sort{Ascending: false},
		Filter: request.Filter{WithPhoto: -1, TagIDs: []int{}, TextSearch: "", AllowedCategoriesOnly: true, ObjectIDs: []int{5067}, Brands: []string{}, ImtID: 0},
		Cursor: request.Cursor{Limit: 1500},
	})
	if err != nil {
		s.log.FatalLog("Error updating nomenclatures: %s\n", err)
	}
	s.log.Log(
		"Updated appellations for %d nomenclatures\n"+
			"Updated media for %d nomenclatures\n"+
			"Updated packages for %d nomenclatures", updateAppellation, updateMedia, updatePackages)
	return nil
}
