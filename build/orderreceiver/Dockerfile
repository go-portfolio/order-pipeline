ARG MODE=prod

# ---------- build stage ----------
FROM golang:1.24-alpine as build

# Устанавливаем зависимости
RUN apk add --no-cache git bash ca-certificates curl build-base protobuf

# Ставим плагины для генерации protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Если debug — ставим Delve
ARG MODE=prod
RUN if [ "$MODE" = "debug" ]; then \
      go install github.com/go-delve/delve/cmd/dlv@latest ; \
    fi

WORKDIR /src

# Копируем go.mod и go.sum для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Генерация protobuf
RUN protoc --proto_path=proto \
    --go_out=proto \
    --go-grpc_out=proto \
    proto/order.proto    

# Сборка бинарника всегда (для prod и debug)
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/orderreceiver ./cmd/orderreceiver

# ---------- runtime stage ----------
FROM alpine:3.18

RUN apk add --no-cache ca-certificates bash

WORKDIR /app
ARG MODE=prod

# Копируем бинарник
COPY --from=build /out/orderreceiver /app/orderreceiver

# Копируем единственный скрипт запуска
COPY --from=build /src/scripts/start.sh /app/start.sh
RUN chmod +x /app/start.sh

# Для debug — копируем Delve и исходники
COPY --from=build /go/bin/dlv /dlv
COPY --from=build /src /app/src

EXPOSE 50051
# Delve
EXPOSE 40000   

ENV SERVICE_BINARY=orderreceiver

# ENTRYPOINT всегда один — управляет режимами
ENTRYPOINT ["/app/start.sh"]
