# Сборка проекта для cacheservice
FROM my-go-build-base AS build

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
COPY --from=build /src/scripts/start-cacheservice.sh /app/start.sh

COPY .env /app/.env  

RUN chmod +x /app/start.sh

EXPOSE 50052
ENTRYPOINT ["/app/start.sh"]
