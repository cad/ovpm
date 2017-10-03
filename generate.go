//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --go_out=plugins=grpc:api/pb
//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --grpc-gateway_out=logtostderr=true:api/pb
//go:generate protoc -I api/pb/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --swagger_out=logtostderr=true:template
//go:generate npm install --cwd webui/ovpm/ --prefix webui/ovpm/
//go:generate npm run build --cwd webui/ovpm/ --prefix webui/ovpm/
//go:generate cp webui/ovpm/public/index.html template/
//go:generate cp webui/ovpm/public/bundle.js template/
//go:generate go-bindata -pkg bindata -o bindata/bindata.go template/

package ovpm
