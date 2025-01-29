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
		case *handlers2.BrandHandler:
			handlerMap["BrandHandler"] = h
		case *handlers2.BarcodeHandler:
			handlerMap["BarcodeHandler"] = h
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

	mux := http.NewServeMux()

	// Проверка и настройка маршрутов для ProductHandler
	if productHandler, ok := handlerMap["ProductHandler"].(*handlers2.ProductHandler); ok {
		mux.HandleFunc("/api/globalids", productHandler.GetGlobalIDsHandler)
		mux.HandleFunc("/api/appellations", productHandler.GetAppellationHandler)
		mux.HandleFunc("/api/descriptions", productHandler.GetDescriptionsHandler)
	} else {
		log.Fatalf("ProductHandler not provided")
	}

	if mediaHandler, ok := handlerMap["MediaHandler"].(*handlers2.MediaHandler); ok {
		mux.HandleFunc("/api/media", mediaHandler.GetMediaHandler)
	} else {
		log.Fatalf("MediaHandler not provided")
	}

	if priceHandler, ok := handlerMap["PriceHandler"].(*handlers2.PriceHandler); ok {
		mux.HandleFunc("/api/price", priceHandler.GetPriceHandler)
	} else {
		log.Fatalf("PriceHandler not provided")
	}

	if sizeHandler, ok := handlerMap["SizeHandler"].(*handlers2.SizeHandler); ok {
		mux.HandleFunc("/api/size", sizeHandler.GetSizeHandler)
	} else {
		log.Fatalf("SizeHandler not provided")
	}

	if brandHandler, ok := handlerMap["BrandHandler"].(*handlers2.BrandHandler); ok {
		mux.HandleFunc("/api/brands", brandHandler.GetBrandHandler)
	} else {
		log.Fatalf("BrandHandler not provided")
	}
	if barcodeHandler, ok := handlerMap["BarcodeHandler"].(*handlers2.BarcodeHandler); ok {
		mux.HandleFunc("/api/barcodes", barcodeHandler.ServeHTTP)
	} else {
		log.Fatalf("BarcodeHandler not provided")
	}

	handlerWithLogging := loggingMiddleware(mux)
	handlerWithCORS := enableCORS(handlerWithLogging)

	log.Printf("Запущен сервис wholesaler /api/")
	log.Fatal(http.ListenAndServe(":8081", handlerWithCORS))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                                // Разрешить запросы со всех источников
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Разрешить методы
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Разрешить заголовки
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
