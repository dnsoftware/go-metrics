package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // для активации декомпрессора
	"google.golang.org/grpc/status"

	"github.com/dnsoftware/go-metrics/internal/constants"
	pb "github.com/dnsoftware/go-metrics/internal/proto"
)

type GRPCServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedMetricsServer

	collector          Collector
	CryptoKey          string
	CertificateKeyPath string
	PrivateKeyPath     string
	TrustedSubnet      string
	Server             *grpc.Server
}

func NewGRPCServer(collector Collector, cryptoKey string, certificateKeyPath string, privateKeyPath string, trustedSubnet string) (*grpc.Server, error) {

	server := &GRPCServer{
		collector:          collector,
		CryptoKey:          cryptoKey,
		CertificateKeyPath: certificateKeyPath,
		PrivateKeyPath:     privateKeyPath,
		TrustedSubnet:      trustedSubnet,
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainUnaryInterceptor(trustedSubnetInterceptor, checkSignInterceptor, loggingInterceptor) /**, grpc.ChainStreamInterceptor(checkSignStreamInterceptor) /**/)

	if certificateKeyPath != "" && privateKeyPath != "" {
		creds, err := credentials.NewServerTLSFromFile(certificateKeyPath, privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("could not load TLS keys for gRPC: %s", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// создаём gRPC-сервер
	server.Server = grpc.NewServer(opts...)

	// регистрируем сервис
	pb.RegisterMetricsServer(server.Server, server)

	return server.Server, nil
}

func (g *GRPCServer) GetMetricValue(ctx context.Context, in *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	var response pb.GetMetricResponse

	if in.MetricType != constants.Gauge && in.MetricType != constants.Counter {
		return nil, status.Errorf(codes.InvalidArgument, `Bad metric type: %s`, in.MetricType)
	}

	val, err := g.collector.GetMetric(ctx, in.MetricType, in.MetricName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, `Metric %s error: %s`, in.MetricName, err.Error())
	}

	response.MetricValue = val

	return &response, nil
}

func (g *GRPCServer) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	var response pb.UpdateMetricResponse

	if in.MetricType != constants.Gauge && in.MetricType != constants.Counter {
		return nil, status.Errorf(codes.InvalidArgument, `Bad metric type: %s`, in.MetricType)
	}

	if in.MetricType == constants.Gauge {
		gaugeVal, err := strconv.ParseFloat(in.MetricValue, 64)

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, `Incorrect metric value: %s`, in.MetricValue)
		}

		err = g.collector.SetGaugeMetric(ctx, in.MetricName, gaugeVal)
		if err != nil {
			return nil, status.Errorf(codes.Internal, `Error when set gauge %s:, %s`, in.MetricName, err.Error())
		}

	}

	if in.MetricType == constants.Counter {
		counterVal, err := strconv.ParseInt(in.MetricValue, 10, 64)

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, `Incorrect metric value: %s`, in.MetricValue)
		}

		err = g.collector.SetCounterMetric(ctx, in.MetricName, counterVal)
		if err != nil {
			return nil, status.Errorf(codes.Internal, `Error when set counter %s:, %s`, in.MetricName, err.Error())
		}

	}

	return &response, nil
}

// GetMetricExt расширенный вариант получения метрики
func (g *GRPCServer) GetMetricExt(ctx context.Context, in *pb.GetMetricExtRequest) (*pb.GetMetricExtResponse, error) {
	var response pb.GetMetricExtResponse
	var err error

	response.Mtype = in.Mtype
	response.Id = in.Id

	if in.Mtype != constants.Gauge && in.Mtype != constants.Counter {
		return nil, status.Errorf(codes.InvalidArgument, `Bad metric type: %s`, in.Mtype)
	}

	switch in.Mtype {
	case constants.Gauge:
		response.Value, err = g.collector.GetGaugeMetric(ctx, in.Id)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, `Error when set gauge %s:, %s`, in.Mtype, err.Error())
		}

	case constants.Counter:
		response.Delta, err = g.collector.GetCounterMetric(ctx, in.Id)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, `Error when set gauge %s:, %s`, in.Mtype, err.Error())
		}
	}

	return &response, nil
}

// UpdateMetricExt расширенный вариант обновления метрики
func (g *GRPCServer) UpdateMetricExt(ctx context.Context, in *pb.UpdateMetricExtRequest) (*pb.UpdateMetricExtResponse, error) {
	var response pb.UpdateMetricExtResponse

	if in.Mtype != constants.Gauge && in.Mtype != constants.Counter {
		return nil, status.Errorf(codes.InvalidArgument, `Bad metric type: %s`, in.Mtype)
	}

	if in.Mtype == constants.Gauge {
		err := g.collector.SetGaugeMetric(ctx, in.Id, in.Value)
		if err != nil {
			return nil, status.Errorf(codes.Internal, `Error when set gauge %s:, %s`, in.Id, err.Error())
		}
	}

	if in.Mtype == constants.Counter {
		err := g.collector.SetCounterMetric(ctx, in.Id, in.Delta)
		if err != nil {
			return nil, status.Errorf(codes.Internal, `Error when set counter %s:, %s`, in.Id, err.Error())
		}

	}

	return &response, nil
}

func (g *GRPCServer) GetAllMetrics(ctx context.Context, in *pb.GetAllMetricsRequest) (*pb.GetAllMetricsResponse, error) {
	var metrics []*pb.GetMetricExtResponse

	gauges, counters, err := g.collector.GetAllByTypes(ctx)
	_, _ = gauges, counters

	if err != nil {
		return nil, status.Errorf(codes.NotFound, `GetAllMetrics error %s`, err.Error())
	}

	for key, val := range gauges {
		metrics = append(metrics, &pb.GetMetricExtResponse{
			Id:    key,
			Mtype: constants.Gauge,
			Value: val,
		})
	}
	for key, val := range counters {
		metrics = append(metrics, &pb.GetMetricExtResponse{
			Id:    key,
			Mtype: constants.Counter,
			Delta: val,
		})
	}

	return &pb.GetAllMetricsResponse{Metrics: metrics}, nil
}

// UpdateMetricsStream Потоковое обновления, двунаправленный поток
func (g *GRPCServer) UpdateMetricsStream(stream pb.Metrics_UpdateMetricsStreamServer) error {

	ctx := context.Background()

	for {
		metric, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		// заносим в базу
		if metric.Mtype != constants.Gauge && metric.Mtype != constants.Counter {
			_ = stream.Send(&pb.UpdateMetricExtResponse{Error: fmt.Sprintf(`Bad metric type: %v, name: %v, value: %v`, metric.Mtype, metric.Id, metric.Value)})
			continue
		}

		if metric.Mtype == constants.Gauge {
			err = g.collector.SetGaugeMetric(ctx, metric.Id, metric.Value)
			if err != nil {
				_ = stream.Send(&pb.UpdateMetricExtResponse{Error: fmt.Sprintf(`SetGaugeMetric error: %v, name: %v, value: %v`, metric.Mtype, metric.Id, metric.Value)})
				continue
			}
		}

		if metric.Mtype == constants.Counter {
			err = g.collector.SetCounterMetric(ctx, metric.Id, metric.Delta)
			if err != nil {
				_ = stream.Send(&pb.UpdateMetricExtResponse{Error: fmt.Sprintf(`SetCounterMetric error: %v, name: %v, value: %v`, metric.Mtype, metric.Id, metric.Delta)})
				continue
			}

		}

		err = stream.Send(&pb.UpdateMetricExtResponse{})
		if err != nil {
			return err
		}
	}

}

func (g *GRPCServer) UpdateMetricsBatch(ctx context.Context, in *pb.UpdateMetricBatchRequest) (*pb.UpdateMetricBatchResponse, error) {
	var response pb.UpdateMetricBatchResponse

	temp := make([]Metrics, 0, len(in.Metrics))
	for _, m := range in.Metrics {
		temp = append(temp, Metrics{
			ID:    m.Id,
			MType: m.Mtype,
			Delta: &m.Delta,
			Value: &m.Value,
		})
	}

	data, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}

	err = g.collector.SetBatchMetrics(ctx, data)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
