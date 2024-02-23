package handlers

import (
	"context"
	"net/http"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/go-chi/chi/v5"
)

type Collector interface {
	SetGaugeMetric(ctx context.Context, name string, value float64) error
	SetCounterMetric(ctx context.Context, name string, value int64) error
	SetBatchMetrics(ctx context.Context, batch []byte) error

	GetGaugeMetric(ctx context.Context, name string) (float64, error)
	GetCounterMetric(ctx context.Context, name string) (int64, error)
	GetMetric(ctx context.Context, metricType string, metricName string) (string, error)
	GetAll(ctx context.Context) (string, error)

	DatabasePing(ctx context.Context) bool
}

type HTTPServer struct {
	collector Collector
	Router    chi.Router
}

// структура для получения json данных от агента
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type (
	// структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// расширенный ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func NewHTTPServer(collector Collector, cryptoKey string) HTTPServer {
	h := HTTPServer{
		collector: collector,
		Router:    NewRouter(),
	}
	h.Router.Use(trimEnd)
	h.Router.Use(CheckSignMiddleware(cryptoKey))
	h.Router.Use(GzipMiddleware)
	h.Router.Use(WithLogging)

	h.Router.Post("/", h.getAllMetrics)
	h.Router.Post("/"+constants.UpdateAction, h.updateMetricJSON)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}", h.noMetricName)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}", h.noMetricValue)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}/{metricValue}", h.updateMetric)
	h.Router.Post("/"+constants.UpdatesAction, h.updatesMetricJSON)

	h.Router.Post("/"+constants.ValueAction, h.getMetricValueJSON)

	h.Router.Get("/", h.getAllMetrics)
	h.Router.Get("/"+constants.ValueAction+"/{metricType}", h.noMetricName)
	h.Router.Get("/"+constants.ValueAction+"/{metricType}/{metricName}", h.getMetricValue)

	h.Router.Get("/ping", h.databasePing)

	return h
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
