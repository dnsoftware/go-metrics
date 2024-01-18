package infrastructure

import (
	"fmt"
	"github.com/dnsoftware/go-metrics/internal/constants"
	"net/http"
)

type Sender interface {
}

type Flags interface {
	RunAddr() string
}

type WebSender struct {
	protocol    string
	domain      string
	contentType string
}

func NewWebSender(protocol string, flags Flags) WebSender {
	return WebSender{
		protocol:    protocol,
		domain:      flags.RunAddr(),
		contentType: "text/plain",
	}
}

func (w *WebSender) SendData(mType string, name string, value string) error {

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
