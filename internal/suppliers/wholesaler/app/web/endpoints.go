package web

import (
	handlers3 "gomarketplace_api/internal/suppliers/wholesaler/app/web/handlers"
	h2 "gomarketplace_api/internal/suppliers/wholesaler/app/web/handlers/h"
	"gomarketplace_api/pkg/middleware"
	"log"
	"net/http"
	"time"
)

// routeConfig хранит конфигурацию маршрута.
type routeConfig struct {
	handlerKey string
	routePath  string
	errMsg     string
	castFunc   func(handlers3.Handler) http.HandlerFunc
}

func SetupRoutes(handlers ...handlers3.Handler) {
	handlerMap := make(map[string]handlers3.Handler)
	for _, handler := range handlers {
		switch h := handler.(type) {
		case *h2.MediaHandler:
			handlerMap["MediaHandler"] = h
		case *h2.PriceHandler:
			handlerMap["PriceHandler"] = h
		case *h2.SizeHandler:
			handlerMap["SizeHandler"] = h
		case *h2.BrandHandler:
			handlerMap["BrandHandler"] = h
		case *h2.BarcodeHandler:
			handlerMap["BarcodeHandler"] = h
		case *h2.AppellationHandler:
			handlerMap["AppellationsHandler"] = h
		case *h2.DescriptionsHandler:
			handlerMap["DescriptionsHandler"] = h
		case *h2.WholesalerIdsHandler:
			handlerMap["IdsHandler"] = h
		default:
			log.Printf("Unknown handler type: %T", h)
		}
	}

	routes := []routeConfig{
		{
			handlerKey: "IdsHandler",
			routePath:  "/api/globalids",
			errMsg:     "WholesalerIdsHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.WholesalerIdsHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
		{
			handlerKey: "AppellationsHandler",
			routePath:  "/api/appellations",
			errMsg:     "AppellationHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.AppellationHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
		{
			handlerKey: "DescriptionsHandler",
			routePath:  "/api/descriptions",
			errMsg:     "DescriptionsHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.DescriptionsHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
		{
			handlerKey: "MediaHandler",
			routePath:  "/api/media",
			errMsg:     "MediaHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.MediaHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
		{
			handlerKey: "PriceHandler",
			routePath:  "/api/price",
			errMsg:     "PriceHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.PriceHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
		{
			handlerKey: "SizeHandler",
			routePath:  "/api/size",
			errMsg:     "SizeHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.SizeHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
		{
			handlerKey: "BrandHandler",
			routePath:  "/api/brands",
			errMsg:     "BrandHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.BrandHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
		{
			handlerKey: "BarcodeHandler",
			routePath:  "/api/barcodes",
			errMsg:     "BarcodeHandler not provided",
			castFunc: func(h handlers3.Handler) http.HandlerFunc {
				handler := h.(*h2.BarcodeHandler)
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		},
	}

	mux := http.NewServeMux()
	for _, rCfg := range routes {
		handler, ok := handlerMap[rCfg.handlerKey]
		if !ok {
			log.Fatalf(rCfg.errMsg)
		}
		mux.Handle(rCfg.routePath, middleware.PrometheusMiddleware(rCfg.castFunc(handler)))
	}

	handlerWithLogging := loggingMiddleware(mux)
	handlerWithCORS := enableCORS(handlerWithLogging)

	log.Printf("Запущен сервис wholesaler на /api/")
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
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}
