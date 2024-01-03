package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (h *HTTPServer) rootHandler(res http.ResponseWriter, req *http.Request) {

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
		}
		return
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
	if parts[2] != "gauge" && parts[2] != "counter" {
		http.Error(res, "Incorrect metric type!", http.StatusBadRequest)
		return
	}

	// имя метрики отсутствует
	if len(parts) == 3 {
		http.Error(res, "Metric name required!", http.StatusNotFound)
		return
	}

	// имя метрики должно быть из списка допустимых
	//...

	// некорректное значение метрики (отсутствует)
	if len(parts) == 4 {
		http.Error(res, "Metric value required!", http.StatusBadRequest)
		return
	}

	// некорректное значение метрики (неправильный тип)
	if parts[2] == "gauge" {
		counterVal, err := strconv.ParseFloat(parts[4], 64)

		if err != nil {
			http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
			return
		}

		fmt.Println(counterVal)
	}

	if parts[2] == "counter" {
		gaugeVal, err := strconv.ParseInt(parts[4], 10, 64)

		if err != nil {
			http.Error(res, "Incorrect metric value!", http.StatusBadRequest)
			return
		}

		fmt.Println(gaugeVal)
	}

	switch parts[1] {
	case "update":

		return

	default:
		h.unrecognized(res, req)
	}

	res.WriteHeader(http.StatusNotFound)
}

func (h *HTTPServer) unrecognized(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Not found!", http.StatusNotFound)
}

func (h *HTTPServer) updateCounter(res http.ResponseWriter, req *http.Request) {

}
