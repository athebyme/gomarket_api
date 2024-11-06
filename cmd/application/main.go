package main

import (
	wsapp "gomarketplace_api/internal/wholesaler/app"
	wbapp "gomarketplace_api/internal/wildberries/app"
	"log"
	"os"
	"sync"
)

func main() {
	log.Printf("\nStarted app\n")
	wg := sync.WaitGroup{}

	synchronize := make(chan struct{}, 1)

	wg.Add(3)
	go func() {
		wsapp.SetupRoutes()
		wg.Done()
	}()
	go func() {
		wserver := wsapp.NewWServer()
		wserver.Run(&synchronize)
		wg.Done()
	}()

	go func() {
		wbserver := wbapp.NewWbServer()
		wbserver.Run(&synchronize)
		wg.Done()
	}()
	wg.Wait()
	os.Exit(0)
}
