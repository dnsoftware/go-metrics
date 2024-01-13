package handlers

import (
	"github.com/go-chi/chi/v5"
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
	//Mux       *http.ServeMux
	collector Collector
	Router    chi.Router
}

func NewHTTPServer(collector Collector) HTTPServer {

	h := HTTPServer{
		//Mux:       http.NewServeMux(),
		collector: collector,
		Router:    NewRouter(),
	}
	h.Router.Use(trimEnd)

	h.Router.Post("/", h.getAllMetrics)
	h.Router.Post("/update", h.noMetricType)
	h.Router.Post("/update/{metricType}", h.noMetricName)
	h.Router.Post("/update/{metricType}/{metricName}", h.noMetricValue)
	h.Router.Post("/update/{metricType}/{metricName}/{metricValue}", h.updateMetric)

	h.Router.Post("/value/{metricType}", h.noMetricName)
	h.Router.Post("/value/{metricType}/{metricName}", h.getMetricValue)

	return h
}
