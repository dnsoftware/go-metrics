package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/dnsoftware/go-metrics/internal/logger"
	"go.uber.org/zap"

	"github.com/dnsoftware/go-metrics/internal/constants"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func trustedSubnetInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	ipStr := GetIP(ctx)
	serv := info.Server.(*GRPCServer)

	if ipStr != "" {
		ipClient := net.ParseIP(ipStr)
		_, ipnet, err := net.ParseCIDR(serv.TrustedSubnet)
		if err != nil {
			return nil, status.Errorf(codes.Internal, `Bad subnet format, %s, %s`, serv.TrustedSubnet, err.Error())
		}
		if !ipnet.Contains(ipClient) {
			return nil, status.Errorf(codes.Unavailable, `Request from untrusted subnet, %s`, serv.TrustedSubnet)
		}
	}

	return handler(ctx, req)
}

func checkSignInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	serv := info.Server.(*GRPCServer)
	serialized, _ := json.Marshal(req)

	if headers, ok := metadata.FromIncomingContext(ctx); ok {
		hashHeader := headers.Get(constants.HashHeaderName)
		if len(hashHeader) > 0 && hashHeader[0] != "" {
			clientHash := hashHeader[0]
			h := hash(serialized, serv.CryptoKey)
			if h != clientHash {
				return nil, status.Errorf(codes.Aborted, `Invalid sign %s`, constants.HashHeaderName)
			}
		}
	}

	return handler(ctx, req)
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	data, _ := json.Marshal(req)

	// отправляем сведения о запросе в лог
	logger.Log().Info("grpc request",
		zap.String("method", info.FullMethod),
		zap.Time("time", time.Now()),
		zap.String("data", string(data)),
	)

	return handler(ctx, req)
}

func GetIP(ctx context.Context) string {
	if headers, ok := metadata.FromIncomingContext(ctx); ok {
		xForwardFor := headers.Get("x-forwarded-for")
		xRealIp := headers.Get("X-Real-IP")
		if len(xForwardFor) > 0 && xForwardFor[0] != "" {
			ips := strings.Split(xForwardFor[0], ",")
			if len(ips) > 0 {
				clientIp := ips[0]
				return clientIp
			}
		} else if len(xRealIp) > 0 && xRealIp[0] != "" {
			clientIp := xRealIp[0]
			return clientIp
		}
	}
	return ""
}

func getLastLineWithSeek(filepath string) string {
	fileHandle, err := os.Open(filepath)

	if err != nil {
		panic("Cannot open file")
		os.Exit(1)
	}
	defer fileHandle.Close()

	line := ""
	var cursor int64 = 0
	stat, _ := fileHandle.Stat()
	filesize := stat.Size()
	for {
		cursor -= 1
		fileHandle.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		fileHandle.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line) // there is more efficient way

		if cursor == -filesize { // stop if we are at the begining
			break
		}
	}

	return line
}
