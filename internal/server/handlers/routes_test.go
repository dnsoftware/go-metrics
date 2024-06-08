package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/storage"
)

func TestHTTPServer_rootHandler(t *testing.T) {
	const (
		ctTextPlain string = "text/plain"
	)

	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name    string
		request string
		method  string
		want    want
	}{
		{
			name:    "only POST negative",
			method:  http.MethodGet,
			request: "/update/",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusMethodNotAllowed,
			},
		},
		{
			name:    "no metric type",
			method:  http.MethodPost,
			request: "/update/",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "bad metric type",
			method:  http.MethodPost,
			request: "/update/badtype",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "no metric name",
			method:  http.MethodPost,
			request: "/update/gauge",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:    "no metric name (with slash)",
			method:  http.MethodPost,
			request: "/update/gauge/",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:    "bad action",
			method:  http.MethodPost,
			request: "/badaction/gauge/metric/val",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:    "incorrect gauge value",
			method:  http.MethodPost,
			request: "/update/gauge/metric/val",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect counter value",
			method:  http.MethodPost,
			request: "/update/counter/metric/val",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect counter value (float)",
			method:  http.MethodPost,
			request: "/update/counter/metric/12.000",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "correct counter value",
			method:  http.MethodPost,
			request: "/update/counter/metric/12",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusOK,
			},
		},
		{
			name:    "correct gauge value",
			method:  http.MethodPost,
			request: "/update/gauge/metric/12.123",
			want: want{
				contentType: ctTextPlain,
				statusCode:  http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.ServerConfig{
				ServerAddress:   "localhost:8080",
				StoreInterval:   constants.BackupPeriod,
				FileStoragePath: constants.FileStoragePath,
				RestoreSaved:    false,
				DatabaseDSN:     "",
			}

			repository := storage.NewMemStorage()
			backupStorage, _ := storage.NewBackupStorage(cfg.FileStoragePath)
			collect, _ := collector.NewCollector(&cfg, repository, backupStorage)
			server := NewServer(collect, "key", nil, "")

			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(server.RootHandler)
			h(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

func TestRouter(t *testing.T) {
	cfg := config.ServerConfig{
		ServerAddress:   "localhost:8080",
		StoreInterval:   constants.BackupPeriod,
		FileStoragePath: constants.FileStoragePath,
		RestoreSaved:    false,
	}

	repository := storage.NewMemStorage()
	backupStorage, _ := storage.NewBackupStorage(cfg.FileStoragePath)
	collect, _ := collector.NewCollector(&cfg, repository, backupStorage)
	server := NewServer(collect, "key", nil, "")
	ts := httptest.NewServer(server.Router)

	postData := "982"

	respPost, _ := testRequest(t, ts, "POST", "/update/counter/testSetGet33/"+postData, nil)
	defer respPost.Body.Close()

	assert.Equal(t, http.StatusOK, respPost.StatusCode)

	respGet, get := testRequest(t, ts, "GET", "/value/counter/testSetGet33", nil)
	defer respGet.Body.Close()

	assert.Equal(t, http.StatusOK, respGet.StatusCode)
	assert.Equal(t, postData, get)

	respGet, _ = testRequest(t, ts, "POST", "/update/bad/testGet33/222", nil)
	defer respGet.Body.Close()
	assert.Equal(t, http.StatusBadRequest, respGet.StatusCode)

	respGet, _ = testRequest(t, ts, "POST", "/update/gauge/testGet33/www", nil)
	defer respGet.Body.Close()
	assert.Equal(t, http.StatusBadRequest, respGet.StatusCode)

	respGet, _ = testRequest(t, ts, "POST", "/update/counter/testGet33/www", nil)
	defer respGet.Body.Close()
	assert.Equal(t, http.StatusBadRequest, respGet.StatusCode)

	respGet, _ = testRequest(t, ts, "GET", "/ping", nil)
	defer respGet.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, respGet.StatusCode)

}

// Тестирование принадлежности IP адреса клиента к подсети
func TestSubnet(t *testing.T) {
	cfg := config.ServerConfig{
		ServerAddress:   "localhost:8080",
		StoreInterval:   constants.BackupPeriod,
		FileStoragePath: constants.FileStoragePath,
		RestoreSaved:    false,
	}

	repository := storage.NewMemStorage()
	backupStorage, _ := storage.NewBackupStorage(cfg.FileStoragePath)
	collect, _ := collector.NewCollector(&cfg, repository, backupStorage)
	server := NewServer(collect, "key", nil, "127.0.0.0/24")
	ts := httptest.NewServer(server.Router)

	headers := make(map[string]string)
	headers[constants.XRealIPName] = "127.0.0.1"

	// позитивный сценарий
	respPost, _ := testRequest(t, ts, "POST", "/update/counter/testSetGet33/111", headers)
	defer respPost.Body.Close()
	assert.Equal(t, http.StatusOK, respPost.StatusCode)

	respGet, _ := testRequest(t, ts, "GET", "/value/counter/testSetGet33", headers)
	defer respGet.Body.Close()
	assert.Equal(t, http.StatusOK, respGet.StatusCode)

	// негативный сценарий
	headers[constants.XRealIPName] = "127.0.1.1"
	respPost, _ = testRequest(t, ts, "POST", "/update/counter/testSetGet33/111", headers)
	defer respPost.Body.Close()
	assert.Equal(t, http.StatusForbidden, respPost.StatusCode)

}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, headers map[string]string) (*http.Response, string) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, nil)
	require.NoError(t, err)

	for i, h := range headers {
		req.Header.Set(i, h)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
