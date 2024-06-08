package infrastructure

import (
	"context"
	"encoding/json"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dnsoftware/go-metrics/internal/constants"
	pb "github.com/dnsoftware/go-metrics/internal/proto"
)

// WebSender отправляет данные на сервер.
type GRPCSender struct {
	domain        string // хост:порт на котором работает gRPC сервер
	cryptoKey     string
	publicKeyPath string            // путь к файлу с публичным асимметричным ключом
	opts          []grpc.DialOption // параметры соединения с сервером

}

// MetricForUnmarshal нужна для конвертации из json со структурным тегом, не совпадающим с именем поля
// как вариант можно тупо править сгенерированный metric.pb.go файл
// или использовать какой-то плагин (с теми что нашел - успеха не добился)
type MetricForUnmarshal struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // структурный тег ЭТОГО поле не совпадает со сгенерированным proto файлом mtype != type
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewGRPCSender(flags Flags, publicKeyPath string) (*GRPCSender, error) {
	var opts []grpc.DialOption

	// используем/не используем ключи шифрования
	if publicKeyPath != "" {
		creds, err := credentials.NewClientTLSFromFile(publicKeyPath, "")
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// ip
	opts = append(opts, grpc.WithUnaryInterceptor(setIPInterceptor))

	// компрессия
	opts = append(opts, grpc.WithUnaryInterceptor(compressInterceptor))

	// хеш подпись отправляемых данных
	if flags.CryptoKey() != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(hashSignInterceptor(flags.CryptoKey())))
	}

	return &GRPCSender{
		domain:        flags.GrpcRunAddr(),
		cryptoKey:     flags.CryptoKey(),
		publicKeyPath: publicKeyPath,
		opts:          opts,
	}, nil
}

// SendData Отправка по одной метрике
func (w *GRPCSender) SendData(ctx context.Context, mType string, name string, value string) error {

	conn, err := grpc.DialContext(ctx, w.domain, w.opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	sendItem := &pb.UpdateMetricExtRequest{
		Mtype: mType,
		Id:    name,
	}

	switch mType {
	case constants.Gauge:
		v, _ := strconv.ParseFloat(value, 64)
		sendItem.Value = v
	case constants.Counter:
		v, _ := strconv.ParseInt(value, 10, 64)
		sendItem.Delta = v
	}

	_, err = client.UpdateMetricExt(ctx, sendItem)

	return err
}

// SendDataBatch отправка данных пакетом
func (w *GRPCSender) SendDataBatch(ctx context.Context, data []byte) error {

	conn, err := grpc.DialContext(ctx, w.domain, w.opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	var metrics []MetricForUnmarshal
	_ = json.Unmarshal(data, &metrics)
	metricsToSend := &pb.UpdateMetricBatchRequest{}

	for _, m := range metrics {
		metricsToSend.Metrics = append(metricsToSend.Metrics, &pb.UpdateMetricExtRequest{
			Id:    m.ID,
			Mtype: m.MType,
			Delta: m.Delta,
			Value: m.Value,
		})
	}

	_, err = client.UpdateMetricsBatch(ctx, metricsToSend)

	return err
}

/*
// SendDataBatchStream отправка данных потоком
func (w *GRPCSender) SendDataBatchStream(ctx context.Context, data []byte) error {

	conn, err := grpc.DialContext(ctx, w.domain, w.opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	stream, err := client.UpdateMetricsStream(ctx)
	if err != nil {
		return err
	}

	var metrics []MetricForUnmarshal
	_ = json.Unmarshal(data, &metrics)

	for _, m := range metrics {
		sendMetric := &pb.UpdateMetricExtRequest{
			Id:    m.ID,
			Mtype: m.MType,
			Delta: m.Delta,
			Value: m.Value,
		}

		_ = stream.Send(sendMetric)
		_, err2 := stream.Recv()
		if err2 == io.EOF {
			break
		}
	}

	return nil
}
*/

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
