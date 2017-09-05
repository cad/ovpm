package api

import (
	"github.com/cad/ovpm/api/pb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	vpnEndPoint     = "localhost:9891" // endpoint of VpnService
	userEndPoint    = "localhost:9892" // endpoint of UserService
	networkEndPoint = "localhost:9893" // endpoint of NetworkService
)

// NewRESTServer returns a new REST server.
func NewRESTServer() (*runtime.ServeMux, context.CancelFunc, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterVPNServiceHandlerFromEndpoint(ctx, mux, vpnEndPoint, opts)
	if err != nil {
		return nil, cancel, err
	}

	err = pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, userEndPoint, opts)
	if err != nil {
		return nil, cancel, err
	}

	err = pb.RegisterNetworkServiceHandlerFromEndpoint(ctx, mux, networkEndPoint, opts)
	if err != nil {
		return nil, cancel, err
	}

	return mux, cancel, nil
}
