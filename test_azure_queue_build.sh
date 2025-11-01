#!/bin/bash
# Azure???????????????????

set -e

cd /home/b0949/develop/grpc-sample-minimal

# ??????
echo "Building and running Azure Queue test program..."
docker compose run --rm -e CGO_ENABLED=1 server sh -c "
    cd /app && 
    go build -o /app/test_azure_queue ./test_azure_queue.go &&
    AZURE_STORAGE_ENDPOINT=http://azurite:10000 \
    AZURE_STORAGE_ACCOUNT_NAME=devstoreaccount1 \
    AZURE_STORAGE_ACCOUNT_KEY=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw== \
    /app/test_azure_queue
"
