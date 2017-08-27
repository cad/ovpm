.PHONY: deps build
docker-build:
	docker run --rm -i -t -e TRAVIS_GO_VERSION=$(TRAVIS_GO_VERSION) -e TRAVIS_BUILD_NUMBER=$(TRAVIS_BUILD_NUMBER) -e TRAVIS_TAG=$(TRAVIS_TAG) -v `pwd`:/fs/src/github.com/cad/ovpm -w /fs/src/github.com/cad/ovpm fedora ./build.sh
	#docker run --rm -i -t -e TRAVIS_GO_VERSION=$(TRAVIS_GO_VERSION) -e TRAVIS_BUILD_NUMBER=$(TRAVIS_BUILD_NUMBER) -e TRAVIS_TAG=$(TRAVIS_TAG) -v `pwd`:/fs/src/github.com/cad/ovpm -w /fs/src/github.com/cad/ovpm fedora /bin/bash


deps:
	# grpc related dependencies
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get -u github.com/golang/protobuf/protoc-gen-go
