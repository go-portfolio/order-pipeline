# Order Processing Pipeline

## Общее описание

`order-pipeline` — это Go-приложение для обработки заказов с использованием пайплайновой архитектуры. Оно построено по принципу событийного взаимодействия между компонентами, что позволяет гибко масштабировать и модифицировать этапы обработки заказа.

## 📌 Архитектура

Поток данных:

Client → gRPC API → Kafka → Worker → Redis → Client

## Микросервисы:

OrderService (gRPC)
Принимает заказы и публикует события в Kafka.

Worker
Подписывается на Kafka, валидирует заказ, выполняет бизнес-логику и сохраняет результат в Redis.

CacheService (gRPC)
Отдаёт результат обработки заказа из Redis для минимальной задержки.

## 🚀 Сборка и запуск
1. Базовый образ

Чтобы не пересобирать тяжёлые зависимости на каждом шаге:

docker build -f build/base/Dockerfile.build-base -t order-build-base .

2. Сборка отдельного микросервиса

Например, для cacheservice:
```
docker compose build cacheservice
docker compose up cacheservice
docker compose exec -it cacheservice sh
```

Пример запуска внутри контейнера:

/app # /app/cacheservice
2025/09/04 06:26:40 .env файл не найден

3. Сборка с docker-compose.override.yml
docker compose up -d --build tests


Для детального вывода логов:
```
DOCKER_BUILDKIT=0 docker compose -f docker-compose.yml -f docker-compose.override.yml build --no-cache --progress=plain tests
```
4. Продакшн

Указать в .env:

MODE=prod


Собрать и запустить:
```
docker compose -f docker-compose.prod.yml up -d orderreceiver --build
```

Проверка состояния контейнеров:
```
docker compose -f docker-compose.prod.yml ps
```
## 🔍 Отладка и тестирование
gRPC-интерфейсы

Получение списка сервисов:
```
grpcurl -plaintext 127.0.0.1:50051 list
```

Результат:
```
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
order.OrderService
```
Тестирование с Delve (dlv)

Запуск в отладочном режиме:

docker exec -it tests bash
dlv test ./tests --headless --listen=:40003 --api-version=2 --accept-multiclient --log

## 🛠️ Используемый стек

Go — микросервисы (gRPC, бизнес-логика)

Kafka — обмен событиями

Redis — кэширование результатов

Docker / Docker Compose — контейнеризация и оркестрация

Delve (dlv) — отладка и тестирование