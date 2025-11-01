#!/bin/bash
# ???????????????????????

set -e

cd /home/b0949/develop/grpc-sample-minimal

# ????????????
echo "Building test program..."
docker compose run --rm -e CGO_ENABLED=1 server sh -c "
    cd /app && 
    go build -o /app/test_queue ./test_queue.go
"

# server????????????????
echo "Running test program..."
docker compose run --rm \
    -e PUBSUB_EMULATOR_HOST=pubsub-emulator:8085 \
    -e GOOGLE_CLOUD_PROJECT=dev-project \
    -e LOCALSTACK_ENDPOINT=http://localstack:4566 \
    -e AWS_ACCESS_KEY_ID=test \
    -e AWS_SECRET_ACCESS_KEY=test \
    -e AWS_REGION=us-east-1 \
    server /app/test_queue
