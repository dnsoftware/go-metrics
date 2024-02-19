package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	return r
}

func (h *HTTPServer) getAllMetrics(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	val, err := h.collector.GetAll(ctx)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}

	res.Header().Set("Content-Type", constants.TextHTML)
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
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	metricType := chi.URLParam(req, constants.MetricType)
	metricName := chi.URLParam(req, constants.MetricName)
	metricValue := chi.URLParam(req, constants.MetricValue)

	if metricType != constants.Gauge && metricType != constants.Counter {
		http.Error(res, "Bad metric type!", http.StatusBadRequest)
		return
	}

	if metricType == constants.Gauge {
		gaugeVal, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
			return
		}

		err = h.collector.SetGaugeMetric(ctx, metricName, gaugeVal)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}

	if metricType == constants.Counter {
		counterVal, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
			return
		}

		err = h.collector.SetCounterMetric(ctx, metricName, counterVal)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

// обновление метрики json формат
func (h *HTTPServer) updateMetricJSON(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	var buf bytes.Buffer

	var metrics Metrics

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if metrics.MType != constants.Gauge && metrics.MType != constants.Counter {
		http.Error(res, "Bad metric type!", http.StatusBadRequest)
		return
	}

	if metrics.MType == constants.Gauge {
		errG := h.collector.SetGaugeMetric(ctx, metrics.ID, *metrics.Value)
		if errG != nil {
			http.Error(res, errG.Error(), http.StatusBadRequest)
			return
		}

		newMetric, errG := h.collector.GetGaugeMetric(ctx, metrics.ID)
		if errG != nil {
			http.Error(res, errG.Error(), http.StatusBadRequest)
			return
		}

		respMetric := Metrics{
			ID:    metrics.ID,
			MType: metrics.MType,
			Value: &newMetric,
		}

		resp, errG := json.Marshal(respMetric)
		if errG != nil {
			http.Error(res, errG.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", constants.ApplicationJSON)
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	}

	if metrics.MType == constants.Counter {
		errC := h.collector.SetCounterMetric(ctx, metrics.ID, *metrics.Delta)
		if errC != nil {
			http.Error(res, errC.Error(), http.StatusBadRequest)
			return
		}

		newMetric, errC := h.collector.GetCounterMetric(ctx, metrics.ID)
		if errC != nil {
			http.Error(res, errC.Error(), http.StatusBadRequest)
			return
		}

		respMetric := Metrics{
			ID:    metrics.ID,
			MType: metrics.MType,
			Delta: &newMetric,
		}

		resp, errC := json.Marshal(respMetric)
		if errC != nil {
			http.Error(res, errC.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", constants.ApplicationJSON)
		res.WriteHeader(http.StatusOK)
		res.Write(resp)
	}
}

// обновление метрик пакетом, json формат
func (h *HTTPServer) updatesMetricJSON(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.collector.SetBatchMetrics(ctx, buf.Bytes())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	res.WriteHeader(http.StatusOK)
}

func (h *HTTPServer) getMetricValue(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	metricType := chi.URLParam(req, constants.MetricType)
	metricName := chi.URLParam(req, constants.MetricName)

	if metricType != constants.Gauge && metricType != constants.Counter {
		http.Error(res, "Bad metric type!", http.StatusBadRequest)
		return
	}

	val, err := h.collector.GetMetric(ctx, metricType, metricName)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(val))
}

func (h *HTTPServer) getMetricValueJSON(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	var buf bytes.Buffer

	var metrics Metrics

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if metrics.MType != constants.Gauge && metrics.MType != constants.Counter {
		http.Error(res, "Bad metric type!", http.StatusBadRequest)
		return
	}

	switch metrics.MType {
	case constants.Gauge:
		val, errG := h.collector.GetGaugeMetric(ctx, metrics.ID)
		if errG != nil {
			http.Error(res, errG.Error(), http.StatusNotFound)
		}

		metrics.Value = &val
	case constants.Counter:
		val, errC := h.collector.GetCounterMetric(ctx, metrics.ID)
		if errC != nil {
			http.Error(res, errC.Error(), http.StatusNotFound)
		}

		metrics.Delta = &val
	}

	resp, err := json.Marshal(metrics)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

// deprecated
// версия из первого инкремента
func (h *HTTPServer) RootHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

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

	// тип метрики отсутствует
	if len(parts) <= 2 {
		http.Error(res, "Incorrect metric type!", http.StatusBadRequest)
		return
	}

	// некорректный тип метрики
	if parts[2] != constants.Gauge && parts[2] != constants.Counter {
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
	case constants.UpdateAction:
		metricType := parts[2]
		metricName := parts[3]

		if metricType == constants.Gauge {
			gaugeVal, err := strconv.ParseFloat(parts[4], 64)

			if err != nil {
				http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
				return
			}

			err = h.collector.SetGaugeMetric(ctx, metricName, gaugeVal)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			res.WriteHeader(http.StatusOK)
		}

		if metricType == constants.Counter {
			counterVal, err := strconv.ParseInt(parts[4], 10, 64)

			if err != nil {
				http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
				return
			}

			err = h.collector.SetCounterMetric(ctx, metricName, counterVal)
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

func (h *HTTPServer) unrecognized(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Not found!", http.StatusNotFound)
}

func (h *HTTPServer) databasePing(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	isConnected := h.collector.DatabasePing(ctx)

	if isConnected {
		res.WriteHeader(http.StatusOK)
	} else {
		res.WriteHeader(http.StatusInternalServerError)
	}
}
