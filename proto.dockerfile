FROM golang:1.18 as builder

RUN apt update && \
    apt install -y protobuf-compiler && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

WORKDIR /go/src/github.com/SphericalPotatoInVacuum/soa-grpc-engine
COPY proto ./proto/
RUN ./proto/generate.sh
