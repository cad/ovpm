package api

import (
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	gcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthUnaryInterceptor is a interceptor function.
//
// See https://godoc.org/google.golang.org/grpc#UnaryServerInterceptor.
func AuthUnaryInterceptor(ctx gcontext.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var enableAuthCheck bool

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("Expected 2 metadata items in context; got %v", md)
	}

	// We enable auth check if we find a non-loopback
	// or invalid IP in the headers coming from the grpc-gateway.
	for _, userAgentIP := range md["x-forwarded-for"] {
		// Check if the remote user IP addr is a proper IP addr.
		if !govalidator.IsIP(userAgentIP) {
			enableAuthCheck = true
			logrus.Debugf("grpc request user agent ip can not be fetched from x-forwarded-for metadata, enabling auth check module '%s'", userAgentIP)
			break
		}

		// Check if the remote user IP addr is a loopback IP addr.
		if ip := net.ParseIP(userAgentIP); !ip.IsLoopback() {
			enableAuthCheck = true
			logrus.Debugf("grpc request user agent ips include non-link local ip, enabling auth check module '%s'", userAgentIP)
			break
		}

		// TODO(cad): We assume gRPC endpoints are for cli only therefore
		//            we are listening only on looback IP.
		//
		// But if we decide use gRPC endpoints publicly, we need to add
		// extra checks against gRPC remote peer IP to test if the request
		// is coming from a remote peer IP or also from a loopback ip.
	}

	if !enableAuthCheck {
		logrus.Debugf("rpc: auth-check not enabled: %s", md["x-forwarded-for"])
	}

	if enableAuthCheck {
		switch info.FullMethod {
		// AuthService methods
		case "/pb.AuthService/Status":
			return authRequired(ctx, req, handler)

		// UserService methods
		case "/pb.UserService/List":
			return authRequired(ctx, req, handler)
		case "/pb.UserService/Create":
			return authRequired(ctx, req, handler)
		case "/pb.UserService/Update":
			return authRequired(ctx, req, handler)
		case "/pb.UserService/Delete":
			return authRequired(ctx, req, handler)
		case "/pb.UserService/Renew":
			return authRequired(ctx, req, handler)
		case "/pb.UserService/GenConfig":
			return authRequired(ctx, req, handler)

		// VPNService methods
		case "/pb.VPNService/Status":
			return authRequired(ctx, req, handler)
		case "/pb.VPNService/Init":
			return authRequired(ctx, req, handler)
		case "/pb.VPNService/Update":
			return authRequired(ctx, req, handler)

		// NetworkService methods
		case "/pb.NetworkService/Create":
			return authRequired(ctx, req, handler)
		case "/pb.NetworkService/List":
			return authRequired(ctx, req, handler)
		case "/pb.NetworkService/Delete":
			return authRequired(ctx, req, handler)
		case "/pb.NetworkService/GetAllTypes":
			return authRequired(ctx, req, handler)
		case "/pb.NetworkService/GetAssociatedUsers":
			return authRequired(ctx, req, handler)
		case "/pb.NetworkService/Associate":
			return authRequired(ctx, req, handler)
		case "/pb.NetworkService/Dissociate":
			return authRequired(ctx, req, handler)
		default:
			logrus.Debugln("rpc: auth is not required for this endpoint: '%s'", info.FullMethod)
		}
	}
	return handler(ctx, req)
}
