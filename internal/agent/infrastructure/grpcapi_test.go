package infrastructure

import (
	"context"
	"crypto/rsa"
	"errors"
	"log"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dnsoftware/go-metrics/internal/logger"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/config"
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"github.com/stretchr/testify/require"
	_ "google.golang.org/grpc/encoding/gzip" // для активации декомпрессора
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 64

var listen *bufconn.Listener

type flags struct {
	cryptoKey   string
	grpcRunAddr string
}

func setup(trustedSubnet string, cryptoSignKey string, certificateKeyPath string, privateKeyPath string) error {
	cfg := config.ServerConfig{
		ServerAddress:   "localhost:8090",
		StoreInterval:   constants.BackupPeriod,
		FileStoragePath: constants.FileStoragePath,
		RestoreSaved:    false,
		DatabaseDSN:     "",
	}

	listen = bufconn.Listen(bufSize)

	repository := storage.NewMemStorage()
	backupStorage, _ := storage.NewBackupStorage(cfg.FileStoragePath)
	collect, _ := collector.NewCollector(&cfg, repository, backupStorage)
	server, err := handlers.NewGRPCServer(collect, cryptoSignKey, certificateKeyPath, privateKeyPath, trustedSubnet)
	if err != nil {
		return errors.New("Not start GRPC server: " + err.Error())
	}

	go func() {
		if err := server.Serve(listen); err != nil {
			log.Fatalf("Test grpc server exited with error: %v", err)
		}
	}()

	return nil
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return listen.Dial()
}

func TestSendData(t *testing.T) {

	cfg := config.ServerConfig{
		ServerAddress:   "localhost:8090",
		StoreInterval:   constants.BackupPeriod,
		FileStoragePath: constants.FileStoragePath,
		RestoreSaved:    false,
		DatabaseDSN:     "",
	}

	lis, _ := net.Listen("tcp", ":8090")

	repository := storage.NewMemStorage()
	backupStorage, _ := storage.NewBackupStorage(cfg.FileStoragePath)
	collect, _ := collector.NewCollector(&cfg, repository, backupStorage)
	server, err := handlers.NewGRPCServer(collect, "qwerty", "", "", "")
	if err != nil {
		logger.Log().Info(err.Error())
	}

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Test grpc server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	flg := &flags{
		cryptoKey:   "qwerty",
		grpcRunAddr: "localhost:8090",
	}

	// отправка метрики
	sender, _ := NewGRPCSender(flg, "")
	testVal := "123.456"
	err = sender.SendData(ctx, constants.Gauge, "Alloc", testVal)
	require.NoError(t, err)

	// отправка пакета
	batch := `[{"id":"Alloc","type":"gauge","value":343728},{"id":"BuckHashSys","type":"gauge","value":7321},{"id":"Frees","type":"gauge","value":268},{"id":"GCCPUFraction","type":"gauge","value":0},{"id":"GCSys","type":"gauge","value":1758064},{"id":"HeapAlloc","type":"gauge","value":343728},{"id":"HeapIdle","type":"gauge","value":2490368},{"id":"HeapInuse","type":"gauge","value":1179648},{"id":"HeapObjects","type":"gauge","value":1959},{"id":"HeapReleased","type":"gauge","value":2490368},{"id":"HeapSys","type":"gauge","value":3670016},{"id":"LastGC","type":"gauge","value":0},{"id":"Lookups","type":"gauge","value":0},{"id":"MCacheInuse","type":"gauge","value":4800},{"id":"MCacheSys","type":"gauge","value":15600},{"id":"MSpanInuse","type":"gauge","value":54400},{"id":"MSpanSys","type":"gauge","value":65280},{"id":"Mallocs","type":"gauge","value":2227},{"id":"NextGC","type":"gauge","value":4194304},{"id":"NumForcedGC","type":"gauge","value":0},{"id":"NumGC","type":"gauge","value":0},{"id":"OtherSys","type":"gauge","value":1126423},{"id":"PauseTotalNs","type":"gauge","value":0},{"id":"StackInuse","type":"gauge","value":524288},{"id":"StackSys","type":"gauge","value":524288},{"id":"Sys","type":"gauge","value":7166992},{"id":"TotalAlloc","type":"gauge","value":343728},{"id":"RandomValue","type":"gauge","value":0.5116380300334399},{"id":"TotalMemory","type":"gauge","value":33518669824},{"id":"FreeMemory","type":"gauge","value":3527917568},{"id":"CPUutilization1","type":"gauge","value":59.793814433156726},{"id":"CPUutilization2","type":"gauge","value":46.487603307343086},{"id":"CPUutilization3","type":"gauge","value":42.47422680367953},{"id":"CPUutilization4","type":"gauge","value":25.63559321940101},{"id":"PollCount","type":"counter","delta":62}]`
	err = sender.SendDataBatch(ctx, []byte(batch))

	require.NoError(t, err)
}

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()
	assert.NotEqual(t, "", ip)
}

func TestNewAgent(t *testing.T) {
	ctx := context.Background()
	_, err := NewAgentRequest(ctx, http.MethodPost, "", []byte("data"), "", &rsa.PublicKey{})
	assert.Error(t, err)

	_, err = NewAgentRequest(ctx, "@#$!`~", "", []byte("data"), "aaa", nil)
	assert.Error(t, err)

	_, err = NewAgentRequest(ctx, http.MethodPost, "", []byte("data"), "xxczxczxc", nil)
	assert.NoError(t, err)

}

func (f *flags) RunAddr() string {
	return ""
}

func (f *flags) CryptoKey() string {
	return f.cryptoKey
}

func (f *flags) ReportInterval() int64 {
	return 10
}

func (f *flags) PollInterval() int64 {
	return 10
}

func (f *flags) RateLimit() int {
	return 10
}

func (f *flags) AsymPubKeyPath() string {
	return ""
}
func (f *flags) GrpcRunAddr() string {
	return f.grpcRunAddr
}
