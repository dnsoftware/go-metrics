package handlers

import (
	"github.com/dnsoftware/go-metrics/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type Middleware func(http.Handler) http.Handler

func trimEnd(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// очистка от конечных пробелов
		r.URL.Path = strings.TrimSpace(r.URL.Path)
		// очистка от конечных слешей
		r.URL.Path = strings.TrimRight(r.URL.Path, "/")

		next.ServeHTTP(w, r)
	})
}

// WithLogging добавляет дополнительный код для регистрации сведений о запросе
// и возвращает новый http.Handler.
func WithLogging(h http.Handler) http.Handler {

	logFn := func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rd := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   rd,
		}

		uri := r.RequestURI
		method := r.Method

		h.ServeHTTP(&lw, r) // внедряем свою реализацию http.ResponseWriter

		// время выполнения запроса.
		duration := time.Since(start)

		// отправляем сведения о запросе в лог
		logger.Log().Info("request",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.Duration("duration", duration),
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
		)
	}

	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}
