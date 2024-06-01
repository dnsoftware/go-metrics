package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/encoding/gzip"

	"google.golang.org/grpc/metadata"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"github.com/dnsoftware/go-metrics/internal/server/config"

	pb "github.com/dnsoftware/go-metrics/internal/proto"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 64

var listen *bufconn.Listener

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
	server, err := NewGRPCServer(collect, cryptoSignKey, certificateKeyPath, privateKeyPath, trustedSubnet)
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

// запрос значения метрики, аналог http getMetricValue
// должен возвратить ошибку, так как база пустая
func TestGetMetricValueNegative(t *testing.T) {
	setup("", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	_, err = client.GetMetricValue(ctx, &pb.GetMetricRequest{
		MetricType: constants.Gauge,
		MetricName: "Alloc",
	})

	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.NotFound)
}

// добавляем/обновляем метрику, аналог http updateMetric
// потом запрашиваем эту метрику и сравниваем значения
// ошибки быть не должно
func TestUpdateAndGetMetric(t *testing.T) {
	setup("", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	// gauge
	// обновление значения метрики, аналог http getMetricValue
	testVal := "123.456"
	respUpd, err := client.UpdateMetric(ctx, &pb.UpdateMetricRequest{
		MetricType:  constants.Gauge,
		MetricName:  "Alloc",
		MetricValue: testVal,
	})

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	// запрос значения метрики, аналог http getMetricValue
	respGet, err := client.GetMetricValue(ctx, &pb.GetMetricRequest{
		MetricType: constants.Gauge,
		MetricName: "Alloc",
	})

	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.OK)
	require.Equal(t, respGet.MetricValue, testVal)

	// counter
	testVal = "123456"
	respUpd, err = client.UpdateMetric(ctx, &pb.UpdateMetricRequest{
		MetricType:  constants.Counter,
		MetricName:  "PollCount",
		MetricValue: testVal,
	})

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	// запрос значения метрики, аналог http getMetricValue
	respGet, err = client.GetMetricValue(ctx, &pb.GetMetricRequest{
		MetricType: constants.Counter,
		MetricName: "PollCount",
	})

	errStatus, ok = status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.OK)
	require.Equal(t, respGet.MetricValue, testVal)

}

// добавляем/обновляем метрику, аналог http updateMetricJSON
// потом запрашиваем эту метрику и сравниваем значения
func TestUpdateAndGetMetricExtended(t *testing.T) {
	setup("", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	// негативный тест - значения нет в базе
	resp2, err := client.GetMetricExt(ctx, &pb.GetMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
	})
	_ = resp2

	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.NotFound)

	// gauge
	// обновление значения метрики, аналог http getMetricValueJSON
	testVal := 123.456
	respUpd, err := client.UpdateMetricExt(ctx, &pb.UpdateMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
		Value: testVal,
	})

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	// запрос значения метрики, аналог http getMetricValueJSON
	respGet, err := client.GetMetricExt(ctx, &pb.GetMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
	})

	errStatus, ok = status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.OK)
	require.Equal(t, testVal, respGet.Value)

	// counter
	testValCounter := int64(123456)
	respUpd, err = client.UpdateMetricExt(ctx, &pb.UpdateMetricExtRequest{
		Mtype: constants.Counter,
		Id:    "PollCount",
		Delta: testValCounter,
	})

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	respGet, err = client.GetMetricExt(ctx, &pb.GetMetricExtRequest{
		Mtype: constants.Counter,
		Id:    "PollCount",
	})

	errStatus, ok = status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.OK)
	require.Equal(t, respGet.Delta, testValCounter)

}

// Получение потока обновлений / отправка потока ответов
func TestUpdateAndGetMetricStream(t *testing.T) {
	setup("", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	stream, err := client.UpdateMetricsStream(ctx)
	require.NoError(t, err)

	batch := `[{"id":"Alloc","mtype":"gauge","value":343728},{"id":"BuckHashSys","mtype":"gauge","value":7321},{"id":"Frees","mtype":"gauge","value":268},{"id":"GCCPUFraction","mtype":"gauge","value":0},{"id":"GCSys","mtype":"gauge","value":1758064},{"id":"HeapAlloc","mtype":"gauge","value":343728},{"id":"HeapIdle","mtype":"gauge","value":2490368},{"id":"HeapInuse","mtype":"gauge","value":1179648},{"id":"HeapObjects","mtype":"gauge","value":1959},{"id":"HeapReleased","mtype":"gauge","value":2490368},{"id":"HeapSys","mtype":"gauge","value":3670016},{"id":"LastGC","mtype":"gauge","value":0},{"id":"Lookups","mtype":"gauge","value":0},{"id":"MCacheInuse","mtype":"gauge","value":4800},{"id":"MCacheSys","mtype":"gauge","value":15600},{"id":"MSpanInuse","mtype":"gauge","value":54400},{"id":"MSpanSys","mtype":"gauge","value":65280},{"id":"Mallocs","mtype":"gauge","value":2227},{"id":"NextGC","mtype":"gauge","value":4194304},{"id":"NumForcedGC","mtype":"gauge","value":0},{"id":"NumGC","mtype":"gauge","value":0},{"id":"OtherSys","mtype":"gauge","value":1126423},{"id":"PauseTotalNs","mtype":"gauge","value":0},{"id":"StackInuse","mtype":"gauge","value":524288},{"id":"StackSys","mtype":"gauge","value":524288},{"id":"Sys","mtype":"gauge","value":7166992},{"id":"TotalAlloc","mtype":"gauge","value":343728},{"id":"RandomValue","mtype":"gauge","value":0.5116380300334399},{"id":"TotalMemory","mtype":"gauge","value":33518669824},{"id":"FreeMemory","mtype":"gauge","value":3527917568},{"id":"CPUutilization1","mtype":"gauge","value":59.793814433156726},{"id":"CPUutilization2","mtype":"gauge","value":46.487603307343086},{"id":"CPUutilization3","mtype":"gauge","value":42.47422680367953},{"id":"CPUutilization4","mtype":"gauge","value":25.63559321940101},{"id":"PollCount","mtype":"counter","delta":62}]`
	var metrics []*pb.UpdateMetricExtRequest
	err = json.Unmarshal([]byte(batch), &metrics)
	require.NoError(t, err)

	for _, metric := range metrics {
		err = stream.Send(metric)
		require.NoError(t, err)

		resp, err2 := stream.Recv()
		if err2 == io.EOF {
			break
		}
		require.NoError(t, err2)
		require.Equal(t, resp.Error, "")

		log.Printf(resp.String())
	}

	// проверяем наличие внесенных метрик
	m, _ := client.GetMetricExt(ctx, &pb.GetMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
	})
	require.Equal(t, float64(343728), m.Value)

	m, _ = client.GetMetricExt(ctx, &pb.GetMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "CPUutilization4",
	})
	require.Equal(t, 25.63559321940101, m.Value)

	m, _ = client.GetMetricExt(ctx, &pb.GetMetricExtRequest{
		Mtype: constants.Counter,
		Id:    "PollCount",
	})
	require.Equal(t, int64(62), m.Delta)

}

func TestGetAllMetrics(t *testing.T) {
	setup("", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	stream, _ := client.UpdateMetricsStream(ctx)
	batch := `[{"id":"Alloc","mtype":"gauge","value":343728},{"id":"BuckHashSys","mtype":"gauge","value":7321},{"id":"Frees","mtype":"gauge","value":268},{"id":"GCCPUFraction","mtype":"gauge","value":0},{"id":"GCSys","mtype":"gauge","value":1758064},{"id":"HeapAlloc","mtype":"gauge","value":343728},{"id":"HeapIdle","mtype":"gauge","value":2490368},{"id":"HeapInuse","mtype":"gauge","value":1179648},{"id":"HeapObjects","mtype":"gauge","value":1959},{"id":"HeapReleased","mtype":"gauge","value":2490368},{"id":"HeapSys","mtype":"gauge","value":3670016},{"id":"LastGC","mtype":"gauge","value":0},{"id":"Lookups","mtype":"gauge","value":0},{"id":"MCacheInuse","mtype":"gauge","value":4800},{"id":"MCacheSys","mtype":"gauge","value":15600},{"id":"MSpanInuse","mtype":"gauge","value":54400},{"id":"MSpanSys","mtype":"gauge","value":65280},{"id":"Mallocs","mtype":"gauge","value":2227},{"id":"NextGC","mtype":"gauge","value":4194304},{"id":"NumForcedGC","mtype":"gauge","value":0},{"id":"NumGC","mtype":"gauge","value":0},{"id":"OtherSys","mtype":"gauge","value":1126423},{"id":"PauseTotalNs","mtype":"gauge","value":0},{"id":"StackInuse","mtype":"gauge","value":524288},{"id":"StackSys","mtype":"gauge","value":524288},{"id":"Sys","mtype":"gauge","value":7166992},{"id":"TotalAlloc","mtype":"gauge","value":343728},{"id":"RandomValue","mtype":"gauge","value":0.5116380300334399},{"id":"TotalMemory","mtype":"gauge","value":33518669824},{"id":"FreeMemory","mtype":"gauge","value":3527917568},{"id":"CPUutilization1","mtype":"gauge","value":59.793814433156726},{"id":"CPUutilization2","mtype":"gauge","value":46.487603307343086},{"id":"CPUutilization3","mtype":"gauge","value":42.47422680367953},{"id":"CPUutilization4","mtype":"gauge","value":25.63559321940101},{"id":"PollCount","mtype":"counter","delta":62}]`
	var metrics []*pb.UpdateMetricExtRequest
	err = json.Unmarshal([]byte(batch), &metrics)

	for _, metric := range metrics {
		err = stream.Send(metric)
		resp, err2 := stream.Recv()
		if err2 == io.EOF {
			break
		}
		require.NoError(t, err2)
		require.Equal(t, resp.Error, "")
	}

	all, errAll := client.GetAllMetrics(ctx, &pb.GetAllMetricsRequest{})
	require.NoError(t, errAll)
	require.Greater(t, len(all.Metrics), 30)

	require.NoError(t, errAll)

}

func TestTrustedSubnetInterceptor(t *testing.T) {
	setup("127.0.0.0/24", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	// положительный тест
	md := metadata.New(map[string]string{constants.XRealIPName: "127.0.0.1"})
	ctx = metadata.NewOutgoingContext(ctx, md)
	_, err = client.GetAllMetrics(ctx, &pb.GetAllMetricsRequest{})
	require.NoError(t, err)

	// отрицательный тест
	md = metadata.New(map[string]string{constants.XRealIPName: "127.0.1.1"})
	ctx = metadata.NewOutgoingContext(ctx, md)
	_, err = client.GetAllMetrics(ctx, &pb.GetAllMetricsRequest{})
	require.Error(t, err)

}

func TestCheckSignInterceptor(t *testing.T) {
	setup("", "testkey", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	// позитивный тест: делаем запрос на добавление и запрос на получение с одинаковым хеш-ключом - данные должны совпадать
	testVal := 123.456
	updRequest := &pb.UpdateMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
		Value: testVal,
	}
	serialized, _ := json.Marshal(updRequest)
	hashKeyVal := "testkey"
	h := hash(serialized, hashKeyVal)
	md := metadata.New(map[string]string{constants.HashHeaderName: h})
	ctx = metadata.NewOutgoingContext(ctx, md)
	respUpd, err := client.UpdateMetricExt(ctx, updRequest)

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	// запрос значения метрики
	getRequest := &pb.GetMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
	}
	serialized, _ = json.Marshal(getRequest)
	h = hash(serialized, hashKeyVal)
	md = metadata.New(map[string]string{constants.HashHeaderName: h})
	ctx = metadata.NewOutgoingContext(ctx, md)
	respGet, err := client.GetMetricExt(ctx, getRequest)

	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.OK)
	require.Equal(t, testVal, respGet.Value)

	// негативный тест: меняем ключ хеширования у клиента: запрос должен завершиться с ошибкой с кодом codes.Aborted
	h = hash(serialized, "badkey")
	md = metadata.New(map[string]string{constants.HashHeaderName: h})
	ctx = metadata.NewOutgoingContext(ctx, md)
	_, err = client.GetMetricExt(ctx, getRequest)

	errStatus, ok = status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.Aborted)

}

func TestGzipGrpc(t *testing.T) {
	setup("", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	compressor := grpc.UseCompressor(gzip.Name)

	testVal := 123.456
	respUpd, err := client.UpdateMetricExt(ctx, &pb.UpdateMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
		Value: testVal,
	}, compressor)

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	// запрос значения метрики
	respGet, err := client.GetMetricExt(ctx, &pb.GetMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
	}, compressor)

	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.OK)
	require.Equal(t, testVal, respGet.Value)

}

func TestAsymmetricCryptoGrpc(t *testing.T) {
	certificateKeyPath := "../../crypto/certificate.pem"
	privateKeyPath := "../../crypto/privatekey.pem"

	err := setup("", "", certificateKeyPath, privateKeyPath)
	require.NoError(t, err)

	ctx := context.Background()

	// негативный тест: запускаем клиента без ключа - должна возникнуть ошибка
	conn, err := grpc.DialContext(ctx, "127.0.0.1", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	testVal := "123.456"
	respUpd, err := client.UpdateMetric(ctx, &pb.UpdateMetricRequest{
		MetricType:  constants.Gauge,
		MetricName:  "Alloc",
		MetricValue: testVal,
	})

	require.Error(t, err)

	// позитивный тест: для клиента указываем публичный ключ - должны успешно передать данные и получить корректные данные
	creds, err := credentials.NewClientTLSFromFile(certificateKeyPath, "")
	require.NoError(t, err)

	conn, err = grpc.DialContext(ctx, "127.0.0.1", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(creds))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client = pb.NewMetricsClient(conn)

	testVal = "123.456"
	respUpd, err = client.UpdateMetric(ctx, &pb.UpdateMetricRequest{
		MetricType:  constants.Gauge,
		MetricName:  "Alloc",
		MetricValue: testVal,
	})

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	// запрос значения метрики, аналог http getMetricValue
	respGet, err := client.GetMetricValue(ctx, &pb.GetMetricRequest{
		MetricType: constants.Gauge,
		MetricName: "Alloc",
	})

	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, errStatus.Code(), codes.OK)
	require.Equal(t, respGet.MetricValue, testVal)

}

func TestLoggingInterceptor(t *testing.T) {
	setup("", "", "", "")
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial : %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	testVal := 12321.45654
	respUpd, err := client.UpdateMetricExt(ctx, &pb.UpdateMetricExtRequest{
		Mtype: constants.Gauge,
		Id:    "Alloc",
		Value: testVal,
	})

	require.NotNil(t, respUpd)
	require.NoError(t, err)

	line := getLastLineWithSeek(constants.LogFile)

	findStr := fmt.Sprintf("%v", testVal)
	require.Contains(t, line, findStr)

}
