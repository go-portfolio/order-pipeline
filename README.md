# План проекта
Проект реализует пайплайн обработки заказов с использованием микросервисной архитектуры:

API-сервис (gRPC) — принимает заказ от клиента.

Publisher — публикует событие «новый заказ» в Kafka.

Worker — подписан на топик Kafka, обрабатывает заказы (валидация, бизнес-логика, расчёт).

Cache Updater — сохраняет результат обработки в Redis.

Client — запрашивает готовые данные из кэша для минимальной задержки.

То есть поток данных:

Client → gRPC API → Kafka → Worker → Redis → Client

# Микросервисный пайплайн заказов:

OrderService (gRPC) → принимает заказы и пишет их в Kafka

Worker → читает заказы из Kafka, обрабатывает, кладёт результат в Redis

CacheService (gRPC) → отдаёт результат заказа из Redis

# Сборка базового образа
Нужно чтобы большие компиляции не повторять каждый раз затрачивая время и ресурсы
```bash
$ docker build -f build/base/Dockerfile.build-base -t order-build-base .
```
```
 naming to docker.io/library/my-go-build-base   
```

Сборка отдельного образа для отладки:
```bash
$ docker compose build cacheservice 
```

Запуск для отладки такой:
```bash
$ docker compose up cacheservice
```
```bash
$ docker compose exec -it cacheservice sh
```
Пример запуска для кэша:
```bash
/app # /app/cacheservice
/src/internal/config/config.go
2025/09/04 06:26:40 .env файл не найден
/app # 
```

Сборка контейнеров с режимом debug:
docker compose build --build-arg MODE=debug ordercache

Для вывода всех информативных сообщений в докере во время сборки:
```
$ DOCKER_BUILDKIT=0 docker compose -f docker-compose.yml -f docker-compose.override.yml build --no-cache --progress=plain    tests
```

