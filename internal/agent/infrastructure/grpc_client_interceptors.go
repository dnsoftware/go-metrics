package infrastructure

import (
	"context"
	"encoding/json"

	"github.com/dnsoftware/go-metrics/internal/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

func setIPInterceptor(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// ip клиента
	ip := GetLocalIP()
	md := metadata.New(map[string]string{constants.XRealIPName: ip})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// вызываем RPC-метод
	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func compressInterceptor(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// включаем компрессию
	compressor := grpc.UseCompressor(gzip.Name)
	opts = append(opts, compressor)

	// вызываем RPC-метод
	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func hashSignInterceptor(hashKeyVal string) func(context.Context, string, interface{}, interface{},
	*grpc.ClientConn, grpc.UnaryInvoker, ...grpc.CallOption) error {

	return func(ctx context.Context, method string, req interface{},
		reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		//	request.Header.Set(constants.HashHeaderName, h)

		serialized, _ := json.Marshal(req)
		h := hash(serialized, hashKeyVal)
		md := metadata.New(map[string]string{constants.HashHeaderName: h})
		ctx = metadata.NewOutgoingContext(ctx, md)

		// вызываем RPC-метод
		err := invoker(ctx, method, req, reply, cc, opts...)

		return err
	}

}
