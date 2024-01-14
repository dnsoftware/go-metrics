package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
)

func NewRouter() chi.Router {

	r := chi.NewRouter()
	//	r.Use(middleware.RedirectSlashes)

	return r
}

func (h *HTTPServer) getAllMetrics(res http.ResponseWriter, req *http.Request) {

	val, err := h.collector.GetAll()
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(val))

}

func (h *HTTPServer) noMetricType(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric type required!", http.StatusBadRequest)
}

func (h *HTTPServer) noMetricName(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric name required!", http.StatusNotFound)
}

func (h *HTTPServer) noMetricValue(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric value required!", http.StatusBadRequest)
}

func (h *HTTPServer) updateMetric(res http.ResponseWriter, req *http.Request) {

	metricType := chi.URLParam(req, "metricType")
	metricName := chi.URLParam(req, "metricName")
	metricValue := chi.URLParam(req, "metricValue")

	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Bad metric type!", http.StatusBadRequest)
		return
	}

	if metricType == "gauge" {
		gaugeVal, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
			return
		}

		err = h.collector.SetGaugeMetric(metricName, gaugeVal)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}

	if metricType == "counter" {
		counterVal, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
			return
		}

		err = h.collector.SetCounterMetric(metricName, counterVal)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}

}

func (h *HTTPServer) getMetricValue(res http.ResponseWriter, req *http.Request) {

	metricType := chi.URLParam(req, "metricType")
	metricName := chi.URLParam(req, "metricName")

	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Bad metric type!", http.StatusBadRequest)
		return
	}

	val, err := h.collector.GetMetric(metricType, metricName)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(val))
}

// deprecated
// версия из первого инкремента
func (h *HTTPServer) RootHandler(res http.ResponseWriter, req *http.Request) {

	// only POST
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	// очистка от конечных пробелов
	url := strings.TrimSpace(req.URL.Path)

	// mainpage
	if url == "/" {
		_, err := res.Write([]byte("Mainpage"))
		if err != nil {
			http.Error(res, "Incorrect metric type!", http.StatusInternalServerError)
			return
		}
	}

	// очистка от конечных слешей
	url = strings.TrimRight(url, "/")

	parts := strings.Split(url, "/")
	/**/
	// тип метрики отсутствует
	if len(parts) <= 2 {
		http.Error(res, "Incorrect metric type!", http.StatusBadRequest)
		return
	}

	// некорректный тип метрики
	if parts[2] != "gauge" && parts[2] != "counter" {
		http.Error(res, "Incorrect metric type!", http.StatusBadRequest)
		return
	}

	// имя метрики отсутствует
	if len(parts) == 3 {
		http.Error(res, "Metric name required!", http.StatusNotFound)
		return
	}

	// некорректное значение метрики (отсутствует)
	if len(parts) == 4 {
		http.Error(res, "Metric value required!", http.StatusBadRequest)
		return
	}

	switch parts[1] {
	case "update":

		metricType := parts[2]
		metricName := parts[3]
		if metricType == "gauge" {
			gaugeVal, err := strconv.ParseFloat(parts[4], 64)

			if err != nil {
				http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
				return
			}

			err = h.collector.SetGaugeMetric(metricName, gaugeVal)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			res.WriteHeader(http.StatusOK)
		}

		if metricType == "counter" {
			counterVal, err := strconv.ParseInt(parts[4], 10, 64)

			if err != nil {
				http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
				return
			}

			err = h.collector.SetCounterMetric(metricName, counterVal)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			res.WriteHeader(http.StatusOK)
		}

		return

	default:
		h.unrecognized(res, req)
	}

}

// deprecated
func (h *HTTPServer) unrecognized(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Not found!", http.StatusNotFound)
}
