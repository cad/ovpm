//go:generate go-bindata -pkg bindata -o bindata/bindata.go template/
//go:generate protoc -I pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis pb/user.proto pb/vpn.proto pb/network.proto --go_out=plugins=grpc:pb
//go:generate protoc -I pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis pb/user.proto pb/vpn.proto pb/network.proto --grpc-gateway_out=logtostderr=true:pb
//go:generate protoc -I pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis pb/user.proto pb/vpn.proto pb/network.proto --swagger_out=logtostderr=true:pb

package ovpm
