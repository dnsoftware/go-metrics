package handlers

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//type TrustedSubnetInterceptor struct {
//	subnet string
//}

//func NewTrustedSubnetInterceptor(trustedSubnet string) TrustedSubnetInterceptor {
//	return TrustedSubnetInterceptor{
//		subnet: trustedSubnet,
//	}
//}

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
