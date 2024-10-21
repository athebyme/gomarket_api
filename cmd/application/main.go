package main

import (
	wsapp "gomarketplace_api/internal/wholesaler/app"
	wbapp "gomarketplace_api/internal/wildberries/app"
	"log"
)

func main() {
	log.Printf("\nStarted app\n")
	wserver := wsapp.NewWServer()
	wserver.Run()

	wbserver := wbapp.NewWbServer()
	wbserver.Run()
}
