#!/bin/bash
# ?????????????????????????????

set -e

echo "=== Backend Upload Test ==="

# AUTH_TOKEN???
AUTH_TOKEN=${AUTH_TOKEN:-"test-token"}

# ?????????????????
TEST_FILE="/tmp/test-backend-$(date +%s).png"
echo "Creating test image file: $TEST_FILE"

# 1x1????PNG?????
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" | base64 -d > "$TEST_FILE"

if [ ! -f "$TEST_FILE" ]; then
    echo "ERROR: Failed to create test file"
    exit 1
fi

echo "Test file created: $(ls -lh $TEST_FILE | awk '{print $5}')"

# ???????????
for provider in "gcs" "s3" "azure"; do
    echo ""
    echo "--- Testing $provider provider ---"
    
    FILENAME="images/test-backend-${provider}-$(date +%s).png"
    
    echo "Uploading file to $provider..."
    echo "  Filename: $FILENAME"
    echo "  Provider: $provider"
    
    # ??????
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -X POST \
        -H "Authorization: $AUTH_TOKEN" \
        -F "uploadFile=@$TEST_FILE;filename=$FILENAME" \
        -F "storageProvider=$provider" \
        http://localhost:8080/api/upload-file)
    
    HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')
    
    if [ "$HTTP_STATUS" != "200" ]; then
        echo "ERROR: Upload failed with HTTP status: $HTTP_STATUS"
        echo "Response: $BODY"
        continue
    fi
    
    echo "SUCCESS: File uploaded successfully"
    echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
    
    # ????????
    echo "Waiting 3 seconds for queue processing..."
    sleep 3
    
    # OCR??????????
    echo ""
    echo "=== OCR Service Logs for $provider (last 10 lines) ==="
    docker compose logs --tail=50 ocr-service | grep -E "($provider|$FILENAME|Enqueue|Dequeue|Processing)" | tail -10 || echo "No matching logs found"
done

echo ""
echo "=== Server Logs (OCR queue related, last 20 lines) ==="
docker compose logs --tail=20 server | grep -E "(OCR|Queue|Enqueue)" || echo "No matching logs found"

echo ""
echo "=== Test Completed ==="

# ???????
rm -f "$TEST_FILE"
