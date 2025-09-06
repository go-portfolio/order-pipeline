#!/bin/sh
set -e

echo "Запуск $SERVICE_BINARY в режиме PROD"
exec /app/$SERVICE_BINARY

