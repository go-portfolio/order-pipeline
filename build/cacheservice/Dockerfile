FROM my-go-build-base AS build

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
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/cacheservice ./cmd/cacheservice


# ---------- final stage ----------
FROM alpine:3.18

# Устанавливаем сертификаты для HTTPS
RUN apk add --no-cache ca-certificates

# Создаём рабочую директорию внутри контейнера
WORKDIR /app

# Копируем бинарник из предыдущего stage
COPY --from=build /out/cacheservice /app/cacheservice

# (опционально) копируем .env внутрь контейнера
# COPY .env /app/.env

# Открываем порт для gRPC
EXPOSE 50052

# Запускаем сервис как основной процесс контейнера
#ENTRYPOINT ["/app/cacheservice"]
CMD ["tail", "-f", "/dev/null"]
