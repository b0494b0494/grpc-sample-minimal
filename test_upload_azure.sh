#!/bin/bash
# Azure???????????????????

set -e

echo "=== Azure Queue Upload Integration Test ==="

# AUTH_TOKEN???
AUTH_TOKEN=${AUTH_TOKEN:-"test-token"}

# ??????????????
TEST_FILE="/tmp/test-azure-upload-$(date +%s).png"
echo "Creating test image file: $TEST_FILE"

# 1x1?????PNG?????
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" | base64 -d > "$TEST_FILE"

if [ ! -f "$TEST_FILE" ]; then
    echo "ERROR: Failed to create test file"
    exit 1
fi

echo "Test file created: $(ls -lh $TEST_FILE | awk '{print $5}')"

# Azure??????????
echo ""
echo "--- Testing Azure provider with Queue ---"

FILENAME="images/test-azure-queue-$(date +%s).png"

echo "Uploading file to Azure..."
echo "  Filename: $FILENAME"
echo "  Provider: azure"

# ??????
RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -X POST \
    -H "Authorization: $AUTH_TOKEN" \
    -F "uploadFile=@$TEST_FILE;filename=$FILENAME" \
    -F "storageProvider=azure" \
    http://localhost:8080/api/upload-file)

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" != "200" ]; then
    echo "ERROR: Upload failed with HTTP status: $HTTP_STATUS"
    echo "Response: $BODY"
    rm -f "$TEST_FILE"
    exit 1
fi

echo "SUCCESS: File uploaded successfully"
echo "$BODY" | jq . 2>/dev/null || echo "$BODY"

# ????????
echo ""
echo "Waiting 5 seconds for queue processing..."
sleep 5

# Azure Queue Storage??????
echo ""
echo "=== Azure Queue Service Logs (last 20 lines) ==="
docker compose logs --tail=50 server 2>/dev/null | grep -iE "(azure|queue|enqueue|dequeue)" | tail -20 || echo "No matching logs found"

# OCR?????????
echo ""
echo "=== OCR Service Logs for Azure (last 20 lines) ==="
docker compose logs --tail=50 ocr-service 2>/dev/null | grep -iE "(azure|$FILENAME|Enqueue|Dequeue|Processing)" | tail -20 || echo "No matching logs found"

# ???????????
echo ""
echo "=== Server Logs (OCR queue related, last 30 lines) ==="
docker compose logs --tail=50 server 2>/dev/null | grep -iE "(OCR|Queue|Enqueue|Dequeue|Azure)" | tail -30 || echo "No matching logs found"

# ????????????????????
echo ""
echo "=== Testing Dequeue from Azure Queue ==="
echo "This will be done by the OCR service automatically"

# ???????
echo "Waiting additional 10 seconds for OCR processing..."
sleep 10

# OCR?????
echo ""
echo "=== Checking OCR Results ==="
OCR_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -H "Authorization: $AUTH_TOKEN" \
    "http://localhost:8080/api/list-ocr-results?storageProvider=azure")

OCR_HTTP_STATUS=$(echo "$OCR_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
OCR_BODY=$(echo "$OCR_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$OCR_HTTP_STATUS" = "200" ]; then
    echo "OCR Results retrieved successfully"
    echo "$OCR_BODY" | jq . 2>/dev/null || echo "$OCR_BODY"
else
    echo "Note: Could not retrieve OCR results (HTTP $OCR_HTTP_STATUS)"
fi

echo ""
echo "=== Test Completed ==="

# ???????
rm -f "$TEST_FILE"
