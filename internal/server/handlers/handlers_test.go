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

func ExampleHTTPServer_UpdateMetric_check() {
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status, string(respBody) == val)

	// Output: 200 OK true
}

func ExampleHTTPServer_UpdateMetricJSON() {
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
	server := NewHTTPServer(collect, "key")

	return httptest.NewServer(server.Router)
}
