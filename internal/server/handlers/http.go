package handlers

import (
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"net/http"
)

type HttpServer struct {
	Mux       *http.ServeMux
	collector Collector
}

type Collector interface {
	UpdateGauge(metricName string, metricValue float64)
	UpdateCounter(metricName string, metricValue int64)
}

func NewHttpServer() HttpServer {

	httpServer := HttpServer{
		Mux:       http.NewServeMux(),
		collector: collector.NewCollector(),
	}

	httpServer.Mux.HandleFunc("/", httpServer.rootHandler)

	//httpServer.Mux.HandleFunc("/update/gauge", httpServer.updateGauge)
	////httpServer.Mux.HandleFunc("/update/gauge/", httpServer.updateGauge)
	//
	////httpServer.Mux.HandleFunc("/update/counter", httpServer.updateCounter)
	//httpServer.Mux.HandleFunc("/update/counter/", httpServer.updateCounter)
	//
	//httpServer.Mux.HandleFunc("/update/", httpServer.badMetricType)

	return httpServer
}

//func (h *HttpServer) onlyPostGuard(res http.ResponseWriter, req *http.Request) bool {
//	if req.Method != http.MethodPost {
//		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
//		return false
//	}
//
//	return true
//}
