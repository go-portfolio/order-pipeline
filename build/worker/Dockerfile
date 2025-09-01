FROM golang:1.21-alpine AS build

RUN apk add --no-cache git build-base protobuf ca-certificates curl bash

WORKDIR /src
ENV GOBIN=/go/bin
ENV PATH=$PATH:/go/bin

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN protoc --proto_path=proto --go_out=internal/pb --go-grpc_out=internal/pb proto/order.proto
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /out/worker /usr/local/bin/worker
ENTRYPOINT ["/usr/local/bin/worker"]
