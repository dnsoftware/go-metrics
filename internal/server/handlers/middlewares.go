package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"go.uber.org/zap"
)

type Middleware func(http.Handler) http.Handler

func CheckSignMiddleware(cryptoKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if h := r.Header.Get(constants.HashHeaderName); h != "" {
				// вычитываем тело запроса для проверки подписи, а потом записываем обратно
				var buf bytes.Buffer

				buf.ReadFrom(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))

				hs := hash(buf.Bytes(), cryptoKey)

				if h != hs {
					http.Error(w, "Bad sign", http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func trimEnd(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == constants.PprofAction {
			next.ServeHTTP(w, r)
		}

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

		// вычитываем тело запроса для логирования, а потом записываем обратно
		var buf bytes.Buffer

		buf.ReadFrom(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))

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
			zap.String("body", buf.String()),
		)
	}

	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}

func GzipMiddleware(h http.Handler) http.Handler {

	gzipFn := func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		outWriter := w

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")

		sendsGzip := strings.Contains(contentEncoding, constants.EncodingGzip)
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// внедряем свою реализацию http.ResponseWriter
		h.ServeHTTP(outWriter, r)
	}

	return http.HandlerFunc(gzipFn)
}

func hash(value []byte, key string) string {
	data := append(value, []byte(key)...)
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}
