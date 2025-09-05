FROM golang:1.24
ARG MODE=prod
WORKDIR /app

# Копируем go.mod и go.sum для кэша зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код тестов
COPY . .

RUN echo "tests: MODE=$MODE  — устанавливаем Delve..."; 
# Для debug — ставим Delve
RUN if [ "$MODE" = "debug" ]; then \
      echo "tests: MODE=debug — устанавливаем Delve..."; \
      go install github.com/go-delve/delve/cmd/dlv@latest ; \
    else \
      echo "MODE=$MODE — пропускаем установку Delve"; \
    fi

# Копируем общий скрипт запуска (относительно контекста сборки)
COPY scripts/start.sh /app/start.sh
RUN chmod +x /app/start.sh

# ENV переменные для тестов
ENV SERVICE_BINARY=tests

# CMD — запускаем общий скрипт
ENTRYPOINT ["/app/start.sh"]

