# Стадия сборки проекта
FROM golang:1.24-alpine as build

# Устанавливаем зависимости и плагины
RUN apk add --no-cache git bash ca-certificates curl build-base protobuf \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Опционально рабочий каталог
WORKDIR /src

# Копируем исходники
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Генерация protobuf кода
RUN protoc --proto_path=proto \
    --go_out=proto \
    --go-grpc_out=proto \
    proto/order.proto    

# Сборка статического бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/orderprocessor ./cmd/orderprocessor


# ---------- final stage ----------
FROM alpine:3.18

# Устанавливаем сертификаты для HTTPS
RUN apk add --no-cache ca-certificates

# Создаём рабочую директорию внутри контейнера
WORKDIR /app

# Копируем бинарник из предыдущего stage
COPY --from=build /out/orderprocessor /app/orderprocessor
COPY --from=build /src/scripts/start.sh /app/start.sh

COPY .env /app/.env  

RUN chmod +x /app/start.sh


ENV SERVICE_BINARY=orderprocessor
ENTRYPOINT ["/app/start.sh"]