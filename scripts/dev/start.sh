#!/bin/bash
set -e

echo "Launching Delve headless debugger on port 40000..."
exec dlv debug ./cmd/$SERVICE \
     --headless \
     --listen=:40000 \
     --api-version=2 \
     --accept-multiclient \
     --log
