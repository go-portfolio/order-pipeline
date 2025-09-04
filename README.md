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
$ docker build -f build/base/Dockerfile.build-base -t my-go-build-base .
```
```
 naming to docker.io/library/my-go-build-base   
```

Сборка отдельного образа для отладки:
```bash
$ docker compose build cacheservice 
```