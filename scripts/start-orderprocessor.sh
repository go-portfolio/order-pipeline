#!/bin/sh
set -e

# Загружаем .env, если есть
if [ -f /app/.env ]; then
  export $(grep -v '^#' /app/.env | xargs)
fi

if [ "$MODE" = "debug" ]; then
  echo "Запуск в режиме DEBUG"
  exec /dlv debug --headless --listen=:40000 --api-version=2 --accept-multiclient ./cmd/orderprocessor
else
  echo "Запуск в режиме PROD"
  exec /app/orderprocessor
fi
