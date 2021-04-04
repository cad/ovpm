FROM fedora:latest
LABEL maintainer="Mustafa Arici (mustafa@arici.io)"

# Deps
RUN dnf install -y git make yarnpkg nodejs protobuf-compiler protobuf-static openvpn golang
RUN go get golang.org/dl/go1.16.3
RUN $(go env GOPATH)/bin/go1.16.3 download

RUN $(go env GOPATH)/bin/go1.16.3 install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    $(go env GOPATH)/bin/go1.16.3 install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    $(go env GOPATH)/bin/go1.16.3 install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
    $(go env GOPATH)/bin/go1.16.3 install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && \
    $(go env GOPATH)/bin/go1.16.3 install github.com/kevinburke/go-bindata/go-bindata@latest && \
    $(go env GOPATH)/bin/go1.16.3 install github.com/goreleaser/nfpm/cmd/nfpm@latest

RUN dnf install -y which iptables
RUN echo "alias go=$(go env GOPATH)/bin/go1.16.3" >> /root/.bashrc
RUN echo "export PATH=$PATH:$(go env GOPATH)/bin" >> /root/.bashrc

VOLUME /app

WORKDIR /app