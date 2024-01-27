package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"net/http"
	"strconv"
	"strings"
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

func (w *WebSender) SendData(mType string, name string, value string) error {

	switch w.contentType {
	case constants.TextPlain:
		return w.sendPlain(mType, name, value)
	case constants.ApplicationJson:
		return w.sendJson(mType, name, value)
	}

	return errors.New("bad send data content type")
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

func (w *WebSender) sendJson(mType string, name string, value string) error {

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

	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
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
