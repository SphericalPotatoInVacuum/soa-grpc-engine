FROM proto-builder as builder

WORKDIR /go/src/github.com/SphericalPotatoInVacuum/soa-grpc-engine

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/server ./cmd/server/
RUN CGO_ENABLED=0 go build -o server ./cmd/server/main.go

FROM scratch
WORKDIR /root/
COPY --from=builder /go/src/github.com/SphericalPotatoInVacuum/soa-grpc-engine/server ./
CMD ["./server"]
