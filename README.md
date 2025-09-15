# Order Processing Pipeline

## Общее описание

`order-pipeline` — это Go-приложение для обработки заказов с пайплайновой архитектурой.  
Приложение построено по принципу событийного взаимодействия между компонентами, что позволяет масштабировать и модифицировать этапы обработки заказа без изменения всей системы.

**Основные возможности:**
- Приём заказов через gRPC
- Асинхронная обработка через Kafka
- Кэширование результатов в Redis
- Поддержка масштабирования и отладки через Docker и Delve

---

## 📌 Архитектура

**Поток данных:**
```
Client → gRPC API → Kafka → Worker → Redis → Client
```

**Компоненты:**

| Микросервис    | Описание |
|----------------|----------|
| **OrderService (gRPC)** | Принимает заказы и публикует события в Kafka |
| **Worker** | Подписывается на Kafka, валидирует заказ, выполняет бизнес-логику и сохраняет результат в Redis |
| **CacheService (gRPC)** | Отдаёт результат обработки заказа из Redis для минимальной задержки |

---

## 🚀 Сборка и запуск

### 1. Базовый образ
Чтобы не пересобирать тяжёлые зависимости на каждом шаге:

```bash
docker build -f build/base/Dockerfile.build-base -t order-build-base .
```
2. Сборка отдельного микросервиса
Например, для cacheservice:

```bash
docker compose build cacheservice
docker compose up cacheservice
docker compose exec -it cacheservice sh
```
Запуск внутри контейнера:
```bash
/app # /app/cacheservice
```
3. Сборка с docker-compose.override.yml
```bash
docker compose up -d --build tests
```
Для детального вывода логов:
```bash
DOCKER_BUILDKIT=0 docker compose -f docker-compose.yml -f docker-compose.override.yml build --no-cache --progress=plain tests
```
4. Продакшн
В .env указываем:
```ini
MODE=prod
```
Сборка и запуск сервиса orderreceiver:
```bash
docker compose -f docker-compose.prod.yml up -d orderreceiver --build
```
Проверка состояния контейнеров:
```bash
docker compose -f docker-compose.prod.yml ps
```
## 🔍 Отладка и тестирование
gRPC-интерфейсы
Получение списка сервисов:
```bash
grpcurl -plaintext 127.0.0.1:50051 list
```
Пример результата:
```
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
order.OrderService
```
## Тестирование с Delve (dlv)
Запуск в отладочном режиме:
```bash
docker exec -it tests bash
dlv test ./tests --headless --listen=:40003 --api-version=2 --accept-multiclient --log
```
## 🛠️ Используемый стек
`Go` — микросервисы (gRPC, бизнес-логика)

`Kafka` — обмен событиями

`Redis` — кэширование результатов

`Docker` / `Docker Compose` — контейнеризация и оркестрация

`Delve (dlv)` — отладка и тестирование

## 📦 Примечания
Архитектура позволяет легко масштабировать `Worker` и `CacheService`.

Все сервисы взаимодействуют асинхронно через `Kafka`, что обеспечивает высокую производительность.

`Redis` используется как кэш для ускорения повторных запросов к результатам обработки.

`Docker` и `docker-compose` позволяют быстро разворачивать окружение для тестирования и продакшна.