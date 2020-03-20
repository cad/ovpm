.PHONY: deps build test bundle-webui clean-bundle bundle-swagger proto bundle build
docker-build:
	docker run --rm -i -t -e TRAVIS_GO_VERSION=$(TRAVIS_GO_VERSION) -e TRAVIS_BUILD_NUMBER=$(TRAVIS_BUILD_NUMBER) -e TRAVIS_TAG=$(TRAVIS_TAG) -v `pwd`:/fs/src/github.com/cad/ovpm -w /fs/src/github.com/cad/ovpm cadthecoder/ovpm-builder:latest
docker-build-shell:
	docker run --rm -i -t -e TRAVIS_GO_VERSION=$(TRAVIS_GO_VERSION) -e TRAVIS_BUILD_NUMBER=$(TRAVIS_BUILD_NUMBER) -e TRAVIS_TAG=$(TRAVIS_TAG) -v `pwd`:/fs/src/github.com/cad/ovpm -w /fs/src/github.com/cad/ovpm cadthecoder/ovpm-builder:latest /bin/bash

development-deps:
	# grpc related dependencies
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get -u github.com/golang/protobuf/protoc-gen-go

	# static asset bundling
	go get github.com/kevinburke/go-bindata/...

	# webui related dependencies
	pacman -Sy yarn

test:
	go test -count=1 -race -coverprofile=coverage.txt -covermode=atomic .

proto:
	protoc -I api/pb/ -I$(shell go env GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --go_out=plugins=grpc:api/pb	
	protoc -I api/pb/ -I$(shell go env GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --grpc-gateway_out=logtostderr=true:api/pb

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
	protoc -I api/pb/ -I$(shell go env GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis api/pb/user.proto api/pb/vpn.proto api/pb/network.proto api/pb/auth.proto --swagger_out=logtostderr=true:bundle

bundle: clean-bundle bundle-webui bundle-swagger
	go-bindata -pkg bundle -o bundle/bindata.go bundle/...

build: bundle
	@echo Building
	rm -rf bin/
	mkdir -p bin/
	# CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./bin/ovpm  ./cmd/ovpm 
	# CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./bin/ovpmd ./cmd/ovpmd
	GOOS=linux go build -o ./bin/ovpm  ./cmd/ovpm 
	GOOS=linux go build -o ./bin/ovpmd ./cmd/ovpmd