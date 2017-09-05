//go:generate go-bindata -pkg bindata -o bindata/bindata.go template/
//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto --go_out=plugins=grpc:api/pb
//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto --grpc-gateway_out=logtostderr=true:api/pb
//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto --swagger_out=logtostderr=true:api/pb

package ovpm
