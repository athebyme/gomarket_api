package web

import (
	handlers2 "gomarketplace_api/internal/wholesaler/app/web/handlers"
	"log"
	"net/http"
)

func SetupRoutes(handlers ...handlers2.Handler) {
	// Создаем карту для хранения обработчиков по их типам
	handlerMap := make(map[string]handlers2.Handler)

	// Заполняем карту обработчиков
	for _, handler := range handlers {
		switch h := handler.(type) {
		case *handlers2.ProductHandler:
			handlerMap["ProductHandler"] = h
		case *handlers2.MediaHandler:
			handlerMap["MediaHandler"] = h
		case *handlers2.PriceHandler:
			handlerMap["PriceHandler"] = h
		case *handlers2.SizeHandler:
			handlerMap["SizeHandler"] = h
		default:
			log.Printf("Unknown handler type: %T", h)
		}
	}

	// Проверяем наличие необходимых обработчиков и вызываем Ping для каждого
	for _, handler := range handlerMap {
		if err := handler.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
	}

	// Проверка и настройка маршрутов для ProductHandler
	if productHandler, ok := handlerMap["ProductHandler"].(*handlers2.ProductHandler); ok {
		http.HandleFunc("/api/globalids", productHandler.GetGlobalIDsHandler)
		http.HandleFunc("/api/appellations", productHandler.GetAppellationHandler)
		http.HandleFunc("/api/descriptions", productHandler.GetDescriptionsHandler)
	} else {
		log.Fatalf("ProductHandler not provided")
	}

	if mediaHandler, ok := handlerMap["MediaHandler"].(*handlers2.MediaHandler); ok {
		http.HandleFunc("/api/media", mediaHandler.GetMediaHandler)
	} else {
		log.Fatalf("MediaHandler not provided")
	}

	if priceHandler, ok := handlerMap["PriceHandler"].(*handlers2.PriceHandler); ok {
		http.HandleFunc("/api/price", priceHandler.GetPriceHandler)
	} else {
		log.Fatalf("PriceHandler not provided")
	}

	if sizeHandler, ok := handlerMap["SizeHandler"].(*handlers2.SizeHandler); ok {
		http.HandleFunc("/api/sizes", sizeHandler.GetSizeHandler)
	} else {
		log.Fatalf("SizeHandler not provided")
	}

	log.Printf("Запущен сервис wholesaler /api/")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
