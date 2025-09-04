#!/bin/sh
set -e

# Загружаем .env, если он есть
if [ -f /app/.env ]; then
  export $(grep -v '^#' /app/.env | xargs)
fi

# Запускаем бинарник
exec /app/orderreceiver
