package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "endpoint", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// RecordRequest записывает метрики для HTTP-запроса.
func RecordRequest(method, endpoint string, statusCode int, duration time.Duration) {
	status := classifyStatus(statusCode)
	httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
}

// classifyStatus классифицирует HTTP-статус код в строку.
func classifyStatus(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return "2xx"
	} else if statusCode >= 300 && statusCode < 400 {
		return "3xx"
	} else if statusCode >= 400 && statusCode < 500 {
		return "4xx"
	} else if statusCode >= 500 && statusCode < 600 {
		return "5xx"
	}
	return "unknown"
}

// MetricsHandler возвращает HTTP-обработчик для экспорта метрик Prometheus.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
