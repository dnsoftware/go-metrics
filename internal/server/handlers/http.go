package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/go-chi/chi/v5"

	_ "net/http/pprof"
)

// Collector сборщик метрик. Сохраняет метрики в хранилище. Получает метрики из  хранилища.
type Collector interface {
	// SetGaugeMetric сохранение метрики типа gauge.
	// Параметры: name - название метрики, value - ее значение.
	SetGaugeMetric(ctx context.Context, name string, value float64) error

	// SetCounterMetric сохранение метрики типа counter.
	// Параметры: name - название метрики, value - ее значение.
	SetCounterMetric(ctx context.Context, name string, value int64) error

	// SetBatchMetrics сохраняет метрики в базу пакетом из нескольких штук
	SetBatchMetrics(ctx context.Context, batch []byte) error

	// GetGaugeMetric получение значения метрики типа gauge.
	// Параметры: name - название метрики.
	GetGaugeMetric(ctx context.Context, name string) (float64, error)

	// GetCounterMetric получение значения метрики типа counter из хранилища.
	// Параметры: name - название метрики.
	GetCounterMetric(ctx context.Context, name string) (int64, error)

	// GetMetric получение метрики в текстовом виде
	GetMetric(ctx context.Context, metricType string, metricName string) (string, error)

	// GetAll получение всех метрик списком
	GetAll(ctx context.Context) (string, error)

	// DatabasePing проверка работоспособности СУБД
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
	h.Router.Use(middleware.Compress(5))
	h.Router.Use(WithLogging)

	h.Router.Mount("/debug", middleware.Profiler())

	h.Router.Post("/", h.getAllMetrics)
	h.Router.Post("/"+constants.UpdateAction, h.UpdateMetricJSON)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}", h.noMetricName)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}", h.noMetricValue)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}/{metricValue}", h.UpdateMetric)
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
