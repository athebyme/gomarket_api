package main

import (
	"flag"
	"gomarketplace_api/config"
	wsapp "gomarketplace_api/internal/wholesaler/app"
	wbapp "gomarketplace_api/internal/wildberries/app"
	"gomarketplace_api/pkg/dbconnect/postgres"
	"log"
	"os"
	"runtime"
	"sync"
)

func main() {
	log.Printf("\nStarted app\n")
	runtime.GOMAXPROCS(6)
	pgCfg := config.GetPostgresConfig()
	wbCfg := config.GetWildberriesConfig()

	flag.Parse()

	wg := sync.WaitGroup{}

	synchronize := make(chan struct{}, 1)

	wg.Add(3)
	go func() {
		con := postgres.NewPgConnector(pgCfg)
		handler := wsapp.NewProductHandler(con)
		mediaHandler := wsapp.NewMediaHandler(con)
		wsapp.SetupRoutes(handler, mediaHandler)
		wg.Done()
	}()
	go func() {
		con := postgres.NewPgConnector(pgCfg)
		wserver := wsapp.NewWServer(con)
		wserver.Run(&synchronize)
		wg.Done()
	}()
	go func() {
		con := postgres.NewPgConnector(pgCfg)
		wbserver := wbapp.NewWbServer(con, wbCfg)
		wbserver.Run(&synchronize)
		wg.Done()
	}()
	wg.Wait()
	os.Exit(0)
}
