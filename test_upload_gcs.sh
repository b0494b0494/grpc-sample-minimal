#!/bin/bash
# GCS??????????????PubSub??????????????

set -e

echo "=== GCS Upload and Queue Test ==="

# ??????????????????PNG???
TEST_FILE="/tmp/test-image-$(date +%s).png"
echo "Creating test image file: $TEST_FILE"

# 1x1????PNG??????Base64????????PNG?
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" | base64 -d > "$TEST_FILE"

if [ ! -f "$TEST_FILE" ]; then
    echo "ERROR: Failed to create test file"
    exit 1
fi

echo "Test file created: $(ls -lh $TEST_FILE | awk '{print $5}')"

# ??????namespace????
FILENAME="images/test-upload-$(date +%s).png"

echo ""
echo "Uploading file to GCS..."
echo "  Filename: $FILENAME"
echo "  Provider: gcs"

# WebApp????????????????????
# AUTH_TOKEN????????????????????
AUTH_TOKEN=${AUTH_TOKEN:-"test-token"}

RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -X POST \
    -H "Authorization: $AUTH_TOKEN" \
    -F "uploadFile=@$TEST_FILE;filename=$FILENAME" \
    -F "storageProvider=gcs" \
    http://localhost:8080/api/upload-file)

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

echo ""
echo "Upload Response:"
echo "$BODY" | jq . 2>/dev/null || echo "$BODY"

if [ "$HTTP_STATUS" != "200" ]; then
    echo ""
    echo "ERROR: Upload failed with HTTP status: $HTTP_STATUS"
    exit 1
fi

echo ""
echo "SUCCESS: File uploaded successfully!"
echo ""
echo "Checking OCR service logs for dequeue and processing..."
echo "(Waiting 5 seconds for queue processing)"
sleep 5

echo ""
echo "=== OCR Service Logs (last 30 lines) ==="
docker compose logs --tail=30 ocr-service | grep -E "(GCS|gcs|Dequeue|dequeue|OCR|Processing|Pub/Sub|PubSub)" || echo "No matching logs found"

echo ""
echo "=== Server Logs (OCR queue related, last 20 lines) ==="
docker compose logs --tail=20 server | grep -E "(OCR|Queue|Pub/Sub|PubSub|Enqueue)" || echo "No matching logs found"

echo ""
echo "=== Test Completed ==="
echo "Check the logs above to verify that:"
echo "1. File was uploaded to GCS"
echo "2. OCR task was enqueued to Pub/Sub"
echo "3. OCR service dequeued the task"
echo "4. OCR processing started"

# ???????
rm -f "$TEST_FILE"
