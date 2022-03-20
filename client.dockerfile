FROM proto-builder as builder

WORKDIR /go/src/github.com/SphericalPotatoInVacuum/soa-grpc-engine

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/client ./cmd/client/
RUN CGO_ENABLED=0 go build -o client ./cmd/client/main.go

FROM scratch
WORKDIR /root/
COPY --from=builder /go/src/github.com/SphericalPotatoInVacuum/soa-grpc-engine/client ./
CMD ["./client"]
