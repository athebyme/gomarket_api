package main

import (
	"fmt"
	"gomarketplace_api/config"
	wsapp "gomarketplace_api/internal/wholesaler/app"
	"gomarketplace_api/internal/wholesaler/app/web"
	"gomarketplace_api/internal/wholesaler/app/web/handlers"
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

	wg.Add(1)
	go func() {
		con := postgres.NewPgConnector(pgConfig)
		handler := handlers.NewProductHandler(con)
		mediaHandler := handlers.NewMediaHandler(con)
		priceHandler := handlers.NewPriceHandler(con)
		sizeHandler := handlers.NewSizeHandler(con, writer)
		brandHandler := handlers.NewBrandHandler(con, writer)
		barcodesHandler := handlers.NewBarcodeHandler(con, writer)
		defer wg.Done()
		web.SetupRoutes(handler, mediaHandler, priceHandler, sizeHandler, brandHandler, barcodesHandler)
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
