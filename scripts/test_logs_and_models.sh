#!/bin/bash

# Test script for logs streaming and model download

API_URL="http://localhost:8080/api/v1"

echo "=========================================="
echo "Testing Logs & Models System"
echo "=========================================="

# Test 1: Create a test job
echo -e "\n1. Creating test job..."
JOB_RESPONSE=$(curl -s -X POST "$API_URL/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "dataset_id": 1,
    "configuration": {
      "epochs": 3,
      "base_model": "llama-3.2-3b"
    }
  }')

JOB_ID=$(echo $JOB_RESPONSE | grep -o '"ID":[0-9]*' | grep -o '[0-9]*')
echo "Created job ID: $JOB_ID"

# Test 2: Get job logs (JSON)
echo -e "\n2. Fetching job logs (JSON)..."
curl -s "$API_URL/jobs/$JOB_ID/logs?limit=10" | jq '.'

# Test 3: Stream logs with SSE (run for 10 seconds)
echo -e "\n3. Streaming logs with SSE (10 seconds)..."
timeout 10 curl -N "$API_URL/jobs/$JOB_ID/logs" || true

# Test 4: Wait for job to complete
echo -e "\n4. Waiting for job to complete..."
for i in {1..30}; do
  STATUS=$(curl -s "$API_URL/jobs/$JOB_ID" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
  echo "  Status: $STATUS"
  
  if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
    break
  fi
  
  sleep 2
done

# Test 5: List models
echo -e "\n5. Listing models..."
curl -s "$API_URL/models" | jq '.data[] | {id, name, base_model, status, total_size}'

# Test 6: Get model details
echo -e "\n6. Getting model details..."
MODEL_ID=$(curl -s "$API_URL/models" | jq -r '.data[0].ID // empty')

if [ -n "$MODEL_ID" ]; then
  echo "Model ID: $MODEL_ID"
  curl -s "$API_URL/models/$MODEL_ID" | jq '.'
  
  # Test 7: Download model
  echo -e "\n7. Testing model download..."
  curl -s -I "$API_URL/models/$MODEL_ID/download" | head -n 10
  
  # Uncomment to actually download:
  # curl -O "$API_URL/models/$MODEL_ID/download"
  # unzip -l model-*.zip
else
  echo "No models found"
fi

# Test 8: Create evaluation
echo -e "\n8. Creating evaluation..."
if [ -n "$MODEL_ID" ]; then
  EVAL_RESPONSE=$(curl -s -X POST "$API_URL/models/$MODEL_ID/evaluate" \
    -H "Content-Type: application/json" \
    -d '{
      "test_set_path": "datasets/test_split.json",
      "base_model_name": "llama3.2:3b"
    }')
  
  echo $EVAL_RESPONSE | jq '.'
  
  EVAL_ID=$(echo $EVAL_RESPONSE | jq -r '.evaluation_id // empty')
  
  if [ -n "$EVAL_ID" ]; then
    echo -e "\n9. Getting evaluation status..."
    curl -s "$API_URL/evaluations/$EVAL_ID" | jq '.'
  fi
fi

echo -e "\n=========================================="
echo "Tests completed!"
echo "=========================================="
