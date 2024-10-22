package main

import (
	wsapp "gomarketplace_api/internal/wholesaler/app"
	wbapp "gomarketplace_api/internal/wildberries/app"
	"log"
	"sync"
)

func main() {
	log.Printf("\nStarted app\n")
	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() {
		wserver := wsapp.NewWServer()
		wserver.Run()
		wg.Done()
	}()
	go func() {
		wbserver := wbapp.NewWbServer()
		wbserver.Run()
		wg.Done()
	}()
	wg.Wait()
}
