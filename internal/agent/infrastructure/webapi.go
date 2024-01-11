package infrastructure

import (
	"net/http"
)

type Sender interface {
}

type WebSender struct {
	protocol    string
	domain      string
	contentType string
}

func NewWebSender(protocol string, domain string) WebSender {
	return WebSender{
		protocol:    protocol,
		domain:      domain,
		contentType: "text/plain",
	}
}

func (w *WebSender) SendData(mType string, name string, value string) error {

	url := w.protocol + "://" + w.domain + "/update/" + mType + "/" + name + "/" + value

	resp, err := http.Post(url, w.contentType, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
