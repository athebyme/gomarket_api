package middleware

import (
	"gomarketplace_api/metrics"
	"net/http"
	"time"
)

// responseWriter оборачивает http.ResponseWriter для сохранения кода ответа.
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader перехватывает вызов WriteHeader, сохраняя код ответа.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// PrometheusMiddleware оборачивает HTTP-обработчик для сбора метрик.
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Засекаем время начала обработки запроса.
		start := time.Now()

		// Оборачиваем ResponseWriter для сохранения кода ответа.
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK} // По умолчанию статус 200

		// Вызываем следующий обработчик.
		next.ServeHTTP(rw, r)

		// Вычисляем длительность запроса.
		duration := time.Since(start)

		// Записываем метрики: метод, URL (путь), статус и длительность.
		metrics.RecordRequest(r.Method, r.URL.Path, rw.status, duration)
	})
}
