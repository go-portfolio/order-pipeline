# ---------- build stage ----------
FROM golang:1.24-alpine AS build

# Устанавливаем зависимости
RUN apk add --no-cache git bash ca-certificates curl build-base protobuf

# Рабочий каталог
WORKDIR /src

# GOBIN для go install (не обязательно, если используем готовые бинарники)
ENV GOBIN=/go/bin
ENV PATH=$PATH:/go/bin

# Копируем go.mod и go.sum для кэширования модулей
COPY go.mod go.sum ./

# Скачиваем зависимости отдельно, чтобы кэшировать слой
RUN go mod download

# Копируем весь исходный код
COPY . .

# ---------- Установка protoc-плагинов ----------
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Генерация protobuf кода
RUN protoc --proto_path=proto \
    --go_out=proto \
    --go-grpc_out=proto \
    proto/order.proto

# Сборка статического бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/worker ./cmd/worker

# ---------- final stage ----------
FROM alpine:3.18

# CA-сертификаты для HTTPS
RUN apk add --no-cache ca-certificates

# Копируем готовый бинарник
COPY --from=build /out/worker /usr/local/bin/worker

# Порт gRPC
EXPOSE 50051

# Точка входа
ENTRYPOINT ["/usr/local/bin/worker"]
