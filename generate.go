//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --go_out=plugins=grpc:api/pb
//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --grpc-gateway_out=logtostderr=true:api/pb
//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --swagger_out=logtostderr=true:template
//go:generate go-bindata -pkg bindata -o bindata/bindata.go template/

package ovpm
