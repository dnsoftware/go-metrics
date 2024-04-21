package infrastructure

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
)

// Flags возвращает значения флагов запуска программы
type Flags interface {
	RunAddr() string
	CryptoKey() string
}

// WebSender отправляет данные на сервер.
type WebSender struct {
	protocol    string
	domain      string
	contentType string
	cryptoKey   string
}

// Metrics структура для отправки json данных на сервер
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewWebSender(protocol string, flags Flags, contentType string) WebSender {
	return WebSender{
		protocol:    protocol,
		domain:      flags.RunAddr(),
		contentType: contentType,
		cryptoKey:   flags.CryptoKey(),
	}
}

// SendData Отправка по одной метрике, через url или через json
func (w *WebSender) SendData(ctx context.Context, mType string, name string, value string) error {
	switch w.contentType {
	case constants.TextPlain, constants.TextHTML:
		return w.sendPlain(ctx, mType, name, value)
	case constants.ApplicationJSON:
		return w.sendJSON(ctx, mType, name, value)
	}

	return errors.New("bad send data content type")
}

// SendDataBatch отправка данных пакетом в json формате
func (w *WebSender) SendDataBatch(ctx context.Context, data []byte) error {
	url := w.protocol + "://" + w.domain + "/" + constants.UpdatesAction

	request, err := NewAgentRequest(ctx, http.MethodPost, url, data, w.cryptoKey)
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	request.Header.Set("Content-Type", w.contentType)
	request.Header.Set("Content-Encoding", constants.EncodingGzip)

	err = retryRequest(request)

	return err
}

// sendPlain отправка метрики на сервер простым текстом через url.
func (w *WebSender) sendPlain(ctx context.Context, mType string, name string, value string) error {
	url := w.protocol + "://" + w.domain + "/" + constants.UpdateAction + "/" + mType + "/" + name + "/" + value

	request, err := NewAgentRequest(ctx, http.MethodPost, url, nil, w.cryptoKey)
	if err != nil {
		// обрабатываем ошибку
		logger.Log().Error(err.Error())
		return err
	}

	request.Header.Set("Content-Type", w.contentType)

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

// sendJSON отпрвка данных на сервер в json формате.
func (w *WebSender) sendJSON(ctx context.Context, mType string, name string, value string) error {
	url := w.protocol + "://" + w.domain + "/" + constants.UpdateAction

	data := Metrics{
		ID:    name,
		MType: mType,
	}

	switch mType {
	case constants.Gauge:
		v, _ := strconv.ParseFloat(value, 64)
		data.Value = &v
	case constants.Counter:
		v, _ := strconv.ParseInt(value, 10, 64)
		data.Delta = &v
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	request, err := NewAgentRequest(ctx, http.MethodPost, url, body, w.cryptoKey)
	if err != nil {
		// обрабатываем ошибку
		logger.Log().Error(err.Error())
		return err
	}

	request.Header.Set("Content-Type", w.contentType)
	request.Header.Set("Content-Encoding", constants.EncodingGzip)

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// getGzipReader gzip компрессор входяшего потока байтов
// возврат *bytes.Buffer, реализующего интерфейс io.Reader
func getGzipReader(data []byte) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	zb := gzip.NewWriter(buf)

	_, err := zb.Write(data)
	if err != nil {
		return nil, err
	}

	zb.Close()

	return buf, nil
}

func hash(value []byte, key string) string {
	data := append(value, []byte(key)...)
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}

func NewAgentRequest(ctx context.Context, method, url string, data []byte, cryptoKey string) (*http.Request, error) {
	buf := &bytes.Buffer{}

	var err error
	// gzip сжатие
	if data != nil {
		buf, err = getGzipReader(data)
		if err != nil {
			logger.Log().Error(err.Error())
			return nil, err
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return nil, err
	}

	h := hash(buf.Bytes(), cryptoKey)
	if cryptoKey != "" {
		request.Header.Set(constants.HashHeaderName, h)
	}

	return request, nil
}
