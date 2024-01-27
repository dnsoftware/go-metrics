package handlers

import (
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Collector interface {
	SetGaugeMetric(name string, value float64) error
	GetGaugeMetric(name string) (float64, error)

	SetCounterMetric(name string, value int64) error
	GetCounterMetric(name string) (int64, error)

	GetMetric(metricType string, metricName string) (string, error)
	GetAll() (string, error)
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

func NewHTTPServer(collector Collector) HTTPServer {

	h := HTTPServer{
		collector: collector,
		Router:    NewRouter(),
	}
	h.Router.Use(trimEnd)
	h.Router.Use(WithLogging)

	h.Router.Post("/", h.getAllMetrics)
	h.Router.Post("/"+constants.UpdateAction, h.updateMetricJSON)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}", h.noMetricName)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}", h.noMetricValue)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}/{metricValue}", h.updateMetric)

	h.Router.Post("/"+constants.ValueAction, h.getMetricValueJSON)
	h.Router.Get("/"+constants.ValueAction+"/{metricType}", h.noMetricName)
	h.Router.Get("/"+constants.ValueAction+"/{metricType}/{metricName}", h.getMetricValue)

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
