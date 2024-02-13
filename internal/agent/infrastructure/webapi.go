package infrastructure

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/logger"
	"net/http"
	"strconv"
)

type Flags interface {
	RunAddr() string
}

type WebSender struct {
	protocol    string
	domain      string
	contentType string
}

// структура для отправки json данных на сервер
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
	}
}

// SendData Отправка по одной метрике, через url или через json
func (w *WebSender) SendData(mType string, name string, value string) error {

	switch w.contentType {
	case constants.TextPlain, constants.TextHTML:
		return w.sendPlain(mType, name, value)
	case constants.ApplicationJSON:
		return w.sendJSON(mType, name, value)
	}

	return errors.New("bad send data content type")
}

// SendBatch отправка данных пакетом в json формате
func (w *WebSender) SendDataBatch(data []byte) error {
	url := w.protocol + "://" + w.domain + "/" + constants.UpdatesAction

	// gzip сжатие
	buf, err := w.getGzipReader(data)
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		logger.Log().Error(err.Error())
		return err
	}

	request.Header.Set("Content-Type", w.contentType)
	request.Header.Set("Content-Encoding", constants.EncodingGzip)

	err = retryRequest(request)

	return err
}

func (w *WebSender) sendPlain(mType string, name string, value string) error {

	url := w.protocol + "://" + w.domain + "/" + constants.UpdateAction + "/" + mType + "/" + name + "/" + value

	request, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		// обрабатываем ошибку
		fmt.Println(err)
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

func (w *WebSender) sendJSON(mType string, name string, value string) error {

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

	// gzip сжатие
	buf, err := w.getGzipReader(body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		// обрабатываем ошибку
		fmt.Println(err)
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

// gzip компрессор входяшего потока байтов
// возврат *bytes.Buffer, реализующего интерфейс io.Reader
func (w *WebSender) getGzipReader(data []byte) (*bytes.Buffer, error) {

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write(data)
	if err != nil {
		return nil, err
	}
	zb.Close()

	return buf, nil
}
