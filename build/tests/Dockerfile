ARG MODE=prod
FROM golang:1.24

WORKDIR /app

# Копируем go.mod и go.sum для кэша зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код тестов
COPY . .

# Для debug — ставим Delve
RUN if [ "$MODE" = "debug" ]; then \
      go install github.com/go-delve/delve/cmd/dlv@latest ; \
    fi

# Копируем общий скрипт запуска (относительно контекста сборки)
COPY scripts/start.sh /app/start.sh
RUN chmod +x /app/start.sh

# ENV переменные для тестов
ENV SERVICE_BINARY=tests

# CMD — запускаем общий скрипт
CMD ["/app/start.sh"]
