package handlers

import (
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
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

			repository := storage.NewMemStorage()
			collect := collector.NewCollector(&repository)
			server := NewHTTPServer(collect)

			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(server.RootHandler)
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

		})
	}
}
