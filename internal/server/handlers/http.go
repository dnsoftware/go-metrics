package handlers

import (
	"net/http"
)

type Collector interface {
	SetGaugeMetric(name string, value float64) error
	GetGaugeMetric(name string) (float64, error)

	SetCounterMetric(name string, value int64) error
	GetCounterMetric(name string) (int64, error)
}

type HTTPServer struct {
	Mux       *http.ServeMux
	collector Collector
}

func NewHTTPServer(collector Collector) HTTPServer {

	httpServer := HTTPServer{
		Mux:       http.NewServeMux(),
		collector: collector,
	}

	httpServer.Mux.HandleFunc("/", httpServer.RootHandler)

	return httpServer
}
