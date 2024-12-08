package handlers

import (
	"database/sql"
	"net/http"
)

// прописать хендлеры для каждого эндпоинта !

type Handler interface {
	Connect() (*sql.DB, error)
	Ping() error
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
