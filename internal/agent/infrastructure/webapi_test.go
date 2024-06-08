package infrastructure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/crypto"
)

type fl struct {
	runAddress string
}

func TestWebapi(t *testing.T) {

	router := chi.NewRouter()
	router.Post("/update/gauge/Alloc/123.456", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Post("/updates", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	svr := httptest.NewServer(router)
	defer svr.Close()

	flg := fl{
		runAddress: strings.ReplaceAll(svr.URL, "http://", ""),
	}
	certFilename, _, _ := crypto.DefaultCryptoFilesName()
	publicCryptoKey, _ := crypto.MakePublicKey(certFilename)

	ctx := context.Background()
	sender := NewWebSender("http", &flg, constants.ApplicationJSON, publicCryptoKey)

	err := sender.SendPlain(ctx, constants.Gauge, "Alloc", "123.456")
	assert.NoError(t, err)

	err = sender.SendJSON(ctx, constants.Gauge, "Alloc", "123.456")
	assert.NoError(t, err)

	err = sender.SendJSON(ctx, constants.Counter, "PollCount", "123456")
	assert.NoError(t, err)

	err = sender.SendDataBatch(ctx, []byte("{}"))
	assert.NoError(t, err)

	err = sender.SendData(ctx, constants.Gauge, "Alloc", "123.456")
	assert.NoError(t, err)

	sender = NewWebSender("http", &flg, constants.TextPlain, publicCryptoKey)
	err = sender.SendData(ctx, constants.Gauge, "Alloc", "123.456")
	assert.NoError(t, err)

	sender = NewWebSender("http", &flg, "badtype", publicCryptoKey)
	err = sender.SendData(ctx, constants.Gauge, "Alloc", "123.456")
	assert.Error(t, err)

}

func (f *fl) RunAddr() string {
	return f.runAddress
}

func (f *fl) CryptoKey() string {
	return ""
}

func (f *fl) ReportInterval() int64 {
	return 10
}

func (f *fl) PollInterval() int64 {
	return 10
}

func (f *fl) RateLimit() int {
	return 10
}

func (f *fl) AsymPubKeyPath() string {
	return ""
}
func (f *fl) GrpcRunAddr() string {
	return ""
}
