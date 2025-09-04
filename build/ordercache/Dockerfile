ARG MODE=prod

# ---------- build stage ----------
FROM golang:1.24-alpine as build

# Устанавливаем зависимости
RUN apk add --no-cache git bash ca-certificates curl build-base protobuf

# Ставим плагины для генерации protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Если debug — ставим delve
ARG MODE=prod
RUN if [ "$MODE" = "debug" ]; then \
      go install github.com/go-delve/delve/cmd/dlv@latest ; \
    fi

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Генерация protobuf кода
RUN protoc --proto_path=proto \
    --go_out=proto \
    --go-grpc_out=proto \
    proto/order.proto    

# Сборка бинарника (для prod)
RUN if [ "$MODE" = "prod" ]; then \
      CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/ordercache ./cmd/ordercache ; \
    fi


# ---------- runtime stage ----------
FROM alpine:3.18

RUN apk add --no-cache ca-certificates bash

WORKDIR /app
ARG MODE=prod

# бинарник (только для prod)
COPY --from=build /out/ordercache /app/ordercache

# один общий скрипт запуска
COPY --from=build /src/scripts/start-ordercache.sh /app/start.sh
RUN chmod +x /app/start.sh

# для debug — копируем delve и исходники
COPY --from=build /go/bin/dlv /dlv
COPY --from=build /src /app/src

EXPOSE 50052
# delve
EXPOSE 40000   

ENTRYPOINT ["/app/start.sh"]
