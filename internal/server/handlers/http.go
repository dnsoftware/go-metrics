package handlers

import (
	"github.com/dnsoftware/go-metrics/internal/constants"
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
	collector Collector
	Router    chi.Router
}

func NewHTTPServer(collector Collector) HTTPServer {

	h := HTTPServer{
		collector: collector,
		Router:    NewRouter(),
	}
	h.Router.Use(trimEnd)

	h.Router.Post("/", h.getAllMetrics)
	h.Router.Post("/"+constants.UpdateAction, h.noMetricType)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}", h.noMetricName)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}", h.noMetricValue)
	h.Router.Post("/"+constants.UpdateAction+"/{metricType}/{metricName}/{metricValue}", h.updateMetric)

	h.Router.Get("/"+constants.ValueAction+"/{metricType}", h.noMetricName)
	h.Router.Get("/"+constants.ValueAction+"/{metricType}/{metricName}", h.getMetricValue)

	return h
}
