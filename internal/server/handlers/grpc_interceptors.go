package handlers

import (
	"context"
	"encoding/json"
	"net"
	"strings"

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
