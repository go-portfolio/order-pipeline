#!/bin/bash

set -e

# Ждем, пока OrderService и CacheService будут доступны
echo "Waiting for services..."

while ! nc -z orderservice 50051; do
  echo "Waiting for orderservice..."
  sleep 1
done

while ! nc -z cacheservice 50052; do
  echo "Waiting for cacheservice..."
  sleep 1
done

echo "Services are up. Running e2e tests..."

# Запускаем e2e тесты
go test ./tests/e2e -v
