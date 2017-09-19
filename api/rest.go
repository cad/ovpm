package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm/api/pb"
	"github.com/cad/ovpm/bindata"
	"github.com/go-openapi/runtime/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewRESTServer returns a new REST server.
func NewRESTServer(grpcPort string) (http.Handler, context.CancelFunc, error) {
	mux := http.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	if !govalidator.IsNumeric(grpcPort) {
		return nil, cancel, fmt.Errorf("grpcPort should be numeric")
	}
	endPoint := fmt.Sprintf("localhost:%s", grpcPort)
	ctx = NewOriginTypeContext(ctx, OriginTypeREST)
	gmux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterVPNServiceHandlerFromEndpoint(ctx, gmux, endPoint, opts)
	if err != nil {
		return nil, cancel, err
	}

	err = pb.RegisterUserServiceHandlerFromEndpoint(ctx, gmux, endPoint, opts)
	if err != nil {
		return nil, cancel, err
	}

	err = pb.RegisterNetworkServiceHandlerFromEndpoint(ctx, gmux, endPoint, opts)
	if err != nil {
		return nil, cancel, err
	}

	err = pb.RegisterAuthServiceHandlerFromEndpoint(ctx, gmux, endPoint, opts)
	if err != nil {
		return nil, cancel, err
	}

	mux.HandleFunc("/api/specs/", specsHandler)
	mware := middleware.Redoc(middleware.RedocOpts{
		BasePath: "/api/docs/",
		SpecURL:  "/api/specs/user.swagger.json",
		Path:     "user",
	}, gmux)
	mware = middleware.Redoc(middleware.RedocOpts{
		BasePath: "/api/docs/",
		SpecURL:  "/api/specs/vpn.swagger.json",
		Path:     "vpn",
	}, mware)
	mware = middleware.Redoc(middleware.RedocOpts{
		BasePath: "/api/docs/",
		SpecURL:  "/api/specs/network.swagger.json",
		Path:     "network",
	}, mware)
	mware = middleware.Redoc(middleware.RedocOpts{
		BasePath: "/api/docs/",
		SpecURL:  "/api/specs/auth.swagger.json",
		Path:     "auth",
	}, mware)
	mux.Handle("/api/", mware)
	mux.HandleFunc("/", webuiHandler)

	return allowCORS(mux), cancel, nil
}

func specsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.URL.Path {
	case "/api/specs/user.swagger.json":
		userData, err := bindata.Asset("template/user.swagger.json")
		if err != nil {
			logrus.Warn(err)
		}
		w.Write(userData)

	case "/api/specs/network.swagger.json":
		networkData, err := bindata.Asset("template/network.swagger.json")
		if err != nil {
			logrus.Warn(err)
		}
		w.Write(networkData)
	case "/api/specs/vpn.swagger.json":
		vpnData, err := bindata.Asset("template/vpn.swagger.json")
		if err != nil {
			logrus.Warn(err)
		}
		w.Write(vpnData)
	case "/api/specs/auth.swagger.json":
		vpnData, err := bindata.Asset("template/auth.swagger.json")
		if err != nil {
			logrus.Warn(err)
		}
		w.Write(vpnData)
	}
}

func webuiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/bundle.js":
		userData, err := bindata.Asset("template/bundle.js")
		if err != nil {
			logrus.Warn(err)
		}
		w.Write(userData)

	default:
		networkData, err := bindata.Asset("template/index.html")
		if err != nil {
			logrus.Warn(err)
		}
		w.Write(networkData)
	}
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	logrus.Debugf("rest: preflight request for %s", r.URL.Path)
	return
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
