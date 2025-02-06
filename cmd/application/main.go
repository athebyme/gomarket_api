package main

import (
	"fmt"
	"gomarketplace_api/config"
	wsapp "gomarketplace_api/internal/suppliers/wholesaler/app"
	"gomarketplace_api/internal/suppliers/wholesaler/app/web"
	"gomarketplace_api/internal/suppliers/wholesaler/app/web/handlers/h"
	"gomarketplace_api/internal/suppliers/wholesaler/business"
	"gomarketplace_api/internal/suppliers/wholesaler/storage/repositories"
	wbapp "gomarketplace_api/internal/wildberries/app"
	metrics2 "gomarketplace_api/metrics"
	"gomarketplace_api/pkg/dbconnect/postgres"
	logger2 "gomarketplace_api/pkg/logger"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(6)
	logger := logger2.NewLogger(os.Stdout, "[MainGoroutine]")
	logger.Log("\nStarted app\n")

	appConfig := config.AppConfig{}
	appCfg, err := appConfig.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Log("Config not found or errored. Check config.yaml file !")
		os.Exit(1)
	}

	wbConfig := appCfg.Wildberries
	pgConfig := appCfg.Postgres

	wg := sync.WaitGroup{}

	writer := os.Stdout
	wg.Add(1)

	metrics()

	go func() {
		con := postgres.NewPgConnector(pgConfig)
		wserver := wsapp.NewWServer(con)
		wserver.Run()
		defer wg.Done()
	}()

	wg.Wait()

	wg.Add(2)
	go func() {
		con := postgres.NewPgConnector(pgConfig)
		db, err := con.Connect()
		if err != nil {
			logger.Log("Database not connected")
			os.Exit(1)
		}
		mediaRepo := repositories.NewMediaRepository(db)
		prodRepo := repositories.NewProductRepository(db)
		brandRepo := repositories.NewBrandRepository(prodRepo)
		prodService := business.NewProductService(prodRepo)

		mediaHandler := h.NewMediaHandler(mediaRepo)
		priceHandler := h.NewPriceHandler(db)
		sizeHandler := h.NewSizeHandler(db, writer)
		brandHandler := h.NewBrandHandler(brandRepo)
		barcodesHandler := h.NewBarcodeHandler(db)
		idsHandler := h.NewWholesalerIdsHandler(db)
		appellationsHandler := h.NewAppellationHandler(prodService)
		descriptionsHandler := h.NewDescriptionsHandler(prodService)
		wg.Done()
		web.SetupRoutes(mediaHandler, priceHandler, sizeHandler, brandHandler, barcodesHandler, idsHandler, appellationsHandler, descriptionsHandler)
	}()

	wg.Wait()

	wg.Add(1)
	go func() {
		con := postgres.NewPgConnector(pgConfig)
		wbserver := wbapp.NewWbServer(con, *wbConfig, writer)
		wbserver.Run()
		defer wg.Done()
	}()
	wg.Wait()

	os.Exit(0)
}

func metrics() {
	port := 2112
	http.Handle("/metrics", metrics2.MetricsHandler())
	go func() {
		log.Printf("Starting metrics server on : %d", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
}
