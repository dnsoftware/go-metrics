package infrastructure

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
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
	AsymPubKeyPath() string
}

// WebSender отправляет данные на сервер.
type WebSender struct {
	protocol    string
	domain      string
	contentType string
	cryptoKey   string
	publicKey   *rsa.PublicKey // публичный асимметричный ключ
}

// Metrics структура для отправки json данных на сервер
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewWebSender(protocol string, flags Flags, contentType string, publicKey *rsa.PublicKey) WebSender {

	return WebSender{
		protocol:    protocol,
		domain:      flags.RunAddr(),
		contentType: contentType,
		cryptoKey:   flags.CryptoKey(),
		publicKey:   publicKey,
	}
}

// SendData Отправка по одной метрике, через url или через json
func (w *WebSender) SendData(ctx context.Context, mType string, name string, value string) error {
	switch w.contentType {
	case constants.TextPlain, constants.TextHTML:
		return w.SendPlain(ctx, mType, name, value)
	case constants.ApplicationJSON:
		return w.SendJSON(ctx, mType, name, value)
	}

	return errors.New("bad send data content type")
}

// SendDataBatch отправка данных пакетом в json формате
func (w *WebSender) SendDataBatch(ctx context.Context, data []byte) error {
	url := w.protocol + "://" + w.domain + "/" + constants.UpdatesAction

	request, err := NewAgentRequest(ctx, http.MethodPost, url, data, w.cryptoKey, w.publicKey)
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	request.Header.Set("Content-Type", w.contentType)
	request.Header.Add("Content-Encoding", constants.EncodingGzip)

	err = retryRequest(request)
	//if err != nil {
	//	logger.Log().Error("\n\nDebug batch: " + strconv.Itoa(w.publicKey.Size()) + ", content type: " + request.Header.Get("Content-Encoding"))
	//	logger.Log().Error("Debug retry data: " + string(data) + "\n\n")
	//}

	return err
}

// SendPlain отправка метрики на сервер простым текстом через url.
func (w *WebSender) SendPlain(ctx context.Context, mType string, name string, value string) error {
	url := w.protocol + "://" + w.domain + "/" + constants.UpdateAction + "/" + mType + "/" + name + "/" + value

	request, err := NewAgentRequest(ctx, http.MethodPost, url, nil, w.cryptoKey, w.publicKey)
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

// SendJSON отпрвка данных на сервер в json формате.
func (w *WebSender) SendJSON(ctx context.Context, mType string, name string, value string) error {
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

	request, err := NewAgentRequest(ctx, http.MethodPost, url, body, w.cryptoKey, w.publicKey)
	if err != nil {
		// обрабатываем ошибку
		logger.Log().Error(err.Error())
		return err
	}

	request.Header.Set("Content-Type", w.contentType)
	request.Header.Add("Content-Encoding", constants.EncodingGzip)

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetGzipReader gzip компрессор входяшего потока байтов
// возврат *bytes.Buffer, реализующего интерфейс io.Reader
func GetGzipReader(data []byte) (*bytes.Buffer, error) {
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

func NewAgentRequest(ctx context.Context, method, url string, data []byte, cryptoKey string, publicKey *rsa.PublicKey) (*http.Request, error) {
	buf := &bytes.Buffer{}
	var err error

	// асимметричное шифрование, если нужно
	if publicKey != nil {
		data, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
		if err != nil {
			logger.Log().Error(err.Error())
			return nil, err
		}
	}

	// gzip сжатие
	if data != nil {
		buf, err = GetGzipReader(data)
		if err != nil {
			logger.Log().Error(err.Error())
			return nil, err
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		logger.Log().Error(err.Error())
		return nil, err
	}

	h := hash(buf.Bytes(), cryptoKey)
	if cryptoKey != "" {
		request.Header.Set(constants.HashHeaderName, h)
	}

	// признак асимметричного шифрования
	if publicKey != nil {
		request.Header.Set(constants.CryptoHeaderName, constants.CryptoHeaderValue)
	}

	return request, nil
}
