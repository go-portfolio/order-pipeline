#!/bin/sh
set -e

# Проверяем, что SERVICE_BINARY задан
if [ -z "$SERVICE_BINARY" ]; then
  echo "Ошибка: переменная SERVICE_BINARY не задана!"
  exit 1
fi

# Запуск
if [ "$MODE" = "debug" ]; then
  echo "Запуск $SERVICE_BINARY в режиме DEBUG"
  # Для обычных сервисов
  if [ "$SERVICE_BINARY" != "tests" ]; then
    exec /dlv exec /app/$SERVICE_BINARY --headless --listen=:40000 --api-version=2 --accept-multiclient
  else
    # Для тестов используем dlv test
    exec dlv test ./... --output=/dev/null --headless --listen=:40000 --api-version=2 --accept-multiclient 
  fi
else
  echo "Запуск $SERVICE_BINARY в режиме PROD"
  # Для тестов просто go test, для сервисов бинарник
  if [ "$SERVICE_BINARY" != "tests" ]; then
    exec /app/$SERVICE_BINARY
  else
    go test ./...
  fi
fi
