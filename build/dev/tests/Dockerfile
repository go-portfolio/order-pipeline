# build/tests/Dockerfile
FROM golang:1.24-bullseye

WORKDIR /app

# Устанавливаем зависимости и инструменты
RUN apt-get update && apt-get install -y --no-install-recommends \
    git bash ca-certificates curl build-essential protobuf-compiler \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Копируем go.mod/go.sum и качаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Открываем порт для Delve
EXPOSE 40000

# По умолчанию запускаем контейнер в "спящем" режиме
CMD ["tail", "-f", "/dev/null"]
