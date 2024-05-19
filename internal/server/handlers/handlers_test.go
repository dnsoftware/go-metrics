package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/storage"
)

func ExampleHTTPServer_UpdateMetric_send() {
	ctx := context.Background()

	ts := setupTestServer()

	path := "/update/counter/testUpdate/982"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)

	// Output: 200 OK
}

func ExampleHTTPServer_UpdateMetric_counter_check() {
	ctx := context.Background()

	ts := setupTestServer()

	val := "982"
	path := "/update/counter/testSetGet/" + val
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	path = "/value/counter/testSetGet"
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody) == val)

	// Output: 200 OK true
}

func ExampleHTTPServer_UpdateMetric_gauge_check() {
	ctx := context.Background()

	ts := setupTestServer()

	val := "982"
	path := "/update/gauge/testSetGet/" + val
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	path = "/value/gauge/testSetGet"
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody) == val)

	// Output: 200 OK true
}

func ExampleHTTPServer_getAllMetrics() {
	ctx := context.Background()

	ts := setupTestServer()

	val := "982"
	path := "/update/gauge/testSetGet/" + val
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	path = "/"
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody))

	// Output: 200 OK testSetGet: 982.000000
}

func ExampleHTTPServer_noMetricName() {
	ctx := context.Background()

	ts := setupTestServer()

	path := "/update/gauge"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody))

	// Output: 404 Not Found Metric name required!
}

func ExampleHTTPServer_noMetricValue() {
	ctx := context.Background()

	ts := setupTestServer()

	path := "/update/gauge/Alloc"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody))

	// Output: 400 Bad Request Metric value required!
}

func ExampleHTTPServer_UpdateMetricJSON_gauge() {
	ctx := context.Background()

	ts := setupTestServer()

	path := "/update"
	mVal := float64(123.1)
	m := Metrics{
		ID:    "test",
		MType: "gauge",
		Delta: new(int64),
		Value: &mVal,
	}
	data, _ := json.Marshal(m)
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, buf)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	path = "/value"
	mGet := Metrics{
		ID:    "test",
		MType: "gauge",
	}
	data, _ = json.Marshal(mGet)
	buf = bytes.NewBuffer(data)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, buf)
	if err != nil {
		fmt.Println(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody) == `{"id":"test","type":"gauge","value":123.1}`)

	// Output: 200 OK true
}

func ExampleHTTPServer_UpdateMetricJSON_counter() {
	ctx := context.Background()

	ts := setupTestServer()

	path := "/update"
	mVal := int64(123)
	m := Metrics{
		ID:    "test",
		MType: "counter",
		Delta: &mVal,
		Value: new(float64),
	}
	data, _ := json.Marshal(m)
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, buf)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	path = "/value"
	mGet := Metrics{
		ID:    "test",
		MType: "counter",
	}
	data, _ = json.Marshal(mGet)
	buf = bytes.NewBuffer(data)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, buf)
	if err != nil {
		fmt.Println(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody) == `{"id":"test","type":"counter","delta":123}`)

	// Output: 200 OK true
}

func ExampleHTTPServer_UpdatesMetricJSON() {
	ctx := context.Background()

	ts := setupTestServer()

	path := "/updates"
	mVal := float64(123.1)
	m := Metrics{
		ID:    "test",
		MType: "gauge",
		Delta: new(int64),
		Value: &mVal,
	}

	var metrics []Metrics
	metrics = append(metrics, m)
	data, _ := json.Marshal(metrics)
	buf := bytes.NewBuffer(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, buf)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	path = "/value"
	mGet := Metrics{
		ID:    "test",
		MType: "gauge",
	}
	data, _ = json.Marshal(mGet)
	buf = bytes.NewBuffer(data)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+path, buf)
	if err != nil {
		fmt.Println(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody) == `{"id":"test","type":"gauge","value":123.1}`)

	// Output: 200 OK true
}

func setupTestServer() *httptest.Server {
	cfg := config.ServerConfig{
		ServerAddress:   "localhost:8080",
		StoreInterval:   constants.BackupPeriod,
		FileStoragePath: constants.FileStoragePath,
		RestoreSaved:    false,
	}

	repository := storage.NewMemStorage()
	backupStorage, _ := storage.NewBackupStorage(cfg.FileStoragePath)
	collect, _ := collector.NewCollector(&cfg, repository, backupStorage)
	server := NewServer(collect, "key", nil)

	return httptest.NewServer(server.Router)
}
