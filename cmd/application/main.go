package main

import (
	"gomarketplace_api/config"
	wsapp "gomarketplace_api/internal/wholesaler/app"
	"gomarketplace_api/internal/wholesaler/app/web"
	"gomarketplace_api/internal/wholesaler/app/web/handlers"
	wbapp "gomarketplace_api/internal/wildberries/app"
	"gomarketplace_api/pkg/dbconnect/postgres"
	logger2 "gomarketplace_api/pkg/logger"
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

	synchronize := make(chan struct{}, 1)

	writer := os.Stdout

	wg.Add(3)

	go func() {
		con := postgres.NewPgConnector(pgConfig)
		handler := handlers.NewProductHandler(con)
		mediaHandler := handlers.NewMediaHandler(con)
		priceHandler := handlers.NewPriceHandler(con)
		sizeHandler := handlers.NewSizeHandler(con, writer)
		brandHandler := handlers.NewBrandHandler(con, writer)
		web.SetupRoutes(handler, mediaHandler, priceHandler, sizeHandler, brandHandler)
		wg.Done()
	}()
	go func() {
		con := postgres.NewPgConnector(pgConfig)
		wserver := wsapp.NewWServer(con)
		wserver.Run(&synchronize)
		wg.Done()
	}()
	wg.Wait()
	go func() {
		con := postgres.NewPgConnector(pgConfig)
		wbserver := wbapp.NewWbServer(con, *wbConfig, writer)
		wbserver.Run(&synchronize)
		wg.Done()
	}()
	wg.Wait()
	os.Exit(0)
}
