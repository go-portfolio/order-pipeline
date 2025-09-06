# build/tests/Dockerfile
FROM golang:1.24

WORKDIR /app

# Копируем go.mod и go.sum сначала для кэша зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код тестов
COPY . .

CMD ["tail", "-f", "/dev/null"]

