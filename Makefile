.PHONY: deps build test bundle-webui clean-bundle bundle-swagger proto bundle build

# Runs unit tests.
test:
	go test -count=1 -race -coverprofile=coverage.txt -covermode=atomic .

proto:
	protoc -I./api/pb/ -I/usr/local/include/ --go_opt=paths=source_relative --go_out=./api/pb user.proto vpn.proto network.proto auth.proto
	protoc -I./api/pb/ -I/usr/local/include/ --go-grpc_opt=paths=source_relative --go-grpc_out=./api/pb user.proto vpn.proto network.proto auth.proto
	protoc -I./api/pb/ -I/usr/local/include/ --grpc-gateway_out ./api/pb \
			 --grpc-gateway_opt logtostderr=true \
			 --grpc-gateway_opt paths=source_relative \
			 --grpc-gateway_opt generate_unbound_methods=true \
			 user.proto vpn.proto network.proto auth.proto

clean-bundle:
	@echo Cleaning up bundle/
	rm -rf bundle/
	mkdir -p bundle/

bundle-webui:
	@echo Bundling webui
	yarn --cwd webui/ovpm/ install
	yarn --cwd webui/ovpm/ build 
	cp -r webui/ovpm/build/* bundle

bundle-swagger: proto
	protoc -I./api/pb -I/usr/local/include/ --openapiv2_out=json_names_for_fields=false:./api/pb --openapiv2_opt logtostderr=true user.proto vpn.proto network.proto auth.proto

bundle: clean-bundle bundle-webui bundle-swagger
	go-bindata -pkg bundle -o bundle/bindata.go bundle/...

# Builds server and client binaries under ./bin folder. Accetps $VERSION env var.
build: bundle
	@echo Building
	rm -rf bin/
	mkdir -p bin/
	#CGO_ENABLED=0  GOOS=linux go build -ldflags="-w -X 'github.com/cad/ovpm.Version=$(VERSION)' -extldflags '-static'" -o ./bin/ovpm  ./cmd/ovpm
	#CGO_ENABLED=0  GOOS=linux go build -ldflags="-w -X 'github.com/cad/ovpm.Version=$(VERSION)' -extldflags '-static'" -o ./bin/ovpmd ./cmd/ovpmd

	# Link dynamically for now
	CGO_CFLAGS="-g -O2 -Wno-return-local-addr" go build -ldflags="-X 'github.com/cad/ovpm.Version=$(VERSION)'" -o ./bin/ovpm  ./cmd/ovpm
	CGO_CFLAGS="-g -O2 -Wno-return-local-addr" go build -ldflags="-X 'github.com/cad/ovpm.Version=$(VERSION)'" -o ./bin/ovpmd ./cmd/ovpmd

clean-dist:
	rm -rf dist/
	mkdir -p dist/

# Builds rpm and dep packages under ./dist folder. Accepts $VERSION env var.
dist: clean-dist build
	@echo Generating VERSION=$(VERSION) rpm and deb packages under dist/
	nfpm pkg -t ./dist/ovpm.rpm
	nfpm pkg -t ./dist/ovpm.deb
