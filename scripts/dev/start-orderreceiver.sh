#!/bin/bash
set -e

MODE=${1:-normal}  # "normal" или "debug"
GRPC_PORT=${GRPC_PORT:-50052}


echo "Launching Delve headless debugger on port 40000..."
exec dlv debug ./cmd/orderreceiver \
     --headless \
     --listen=:40000 \
     --api-version=2 \
     --accept-multiclient \
     --log
