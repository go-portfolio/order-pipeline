#!/bin/bash
set -e

MODE=${1:-normal}  # "normal" или "debug"
GRPC_PORT=${GRPC_PORT:-50052}


echo "Launching Delve headless debugger on port 40001..."
exec dlv debug ./cmd/orderprocessor \
     --headless \
     --listen=:40002 \
     --api-version=2 \
     --accept-multiclient \
     --log
