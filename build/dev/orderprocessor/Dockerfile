# ---------- dev/debug stage with Delve ----------
FROM golang:1.24-bullseye

# Устанавливаем зависимости и инструменты + netcat
RUN apt-get update && apt-get install -y --no-install-recommends \
    git bash ca-certificates curl build-essential protobuf-compiler \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем Delve (последняя версия)
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Устанавливаем плагины для protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Рабочая директория
WORKDIR /app

# Копируем go.mod/go.sum и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Генерация protobuf кода
RUN protoc --proto_path=proto \
    --go_out=proto \
    --go-grpc_out=proto \
    proto/order.proto

# Переменные окружения
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPATH=/go \
    PATH=/go/bin:$PATH

# Открываем порты:
# 40002 – Delve debugger
EXPOSE  40002

# Копируем скрипт запуска
COPY scripts/dev/start-orderprocessor.sh /app/start.sh
RUN chmod +x /app/start.sh

# ENTRYPOINT использует скрипт для выбора обычного запуска или Delve
ENTRYPOINT ["/app/start.sh"]
