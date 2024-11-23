package main

import (
	"flag"
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

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Log("Config not found or errored. Check config.yaml file !")
		os.Exit(1)
	}
	pgCfg := cfg.Postgres
	wbCfg := cfg.Wildberries

	flag.Parse()

	wg := sync.WaitGroup{}

	synchronize := make(chan struct{}, 1)

	writer := os.Stdout

	wg.Add(3)
	go func() {
		con := postgres.NewPgConnector(&pgCfg)
		handler := handlers.NewProductHandler(con)
		mediaHandler := handlers.NewMediaHandler(con)
		priceHandler := handlers.NewPriceHandler(con)
		sizeHandler := handlers.NewSizeHandler(con, writer)
		web.SetupRoutes(handler, mediaHandler, priceHandler, sizeHandler)
		wg.Done()
	}()
	go func() {
		con := postgres.NewPgConnector(&pgCfg)
		wserver := wsapp.NewWServer(con)
		wserver.Run(&synchronize)
		wg.Done()
	}()
	wg.Wait()
	go func() {
		con := postgres.NewPgConnector(&pgCfg)
		wbserver := wbapp.NewWbServer(con, wbCfg)
		wbserver.Run(&synchronize)
		wg.Done()
	}()
	wg.Wait()
	os.Exit(0)
}
