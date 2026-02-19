#!/bin/bash

# End-to-End Frontend Testing Script
# Tests complete workflow from UI perspective

set -e

API_URL="http://localhost:8080/api/v1"
FRONTEND_URL="http://localhost:3000"

echo "=========================================="
echo "Frontend End-to-End Testing"
echo "=========================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function
test_endpoint() {
    local name=$1
    local url=$2
    local expected_code=${3:-200}
    
    echo -n "Testing $name... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url")
    
    if [ "$response" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $response)"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_code, got $response)"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test 1: Check services are running
echo -e "\n${YELLOW}1. Checking Services${NC}"
test_endpoint "Backend Health" "$API_URL/health"
test_endpoint "Frontend Home" "$FRONTEND_URL"

# Test 2: Check static files
echo -e "\n${YELLOW}2. Checking Static Files${NC}"
test_endpoint "CSS Stylesheet" "$FRONTEND_URL/css/styles.css"
test_endpoint "API Module" "$FRONTEND_URL/js/api.js"
test_endpoint "Utils Module" "$FRONTEND_URL/js/utils.js"
test_endpoint "Charts Module" "$FRONTEND_URL/js/charts.js"

# Test 3: Check HTML pages
echo -e "\n${YELLOW}3. Checking HTML Pages${NC}"
test_endpoint "Dashboard" "$FRONTEND_URL/index.html"
test_endpoint "Dataset Upload" "$FRONTEND_URL/dataset-upload.html"
test_endpoint "New Training" "$FRONTEND_URL/training-new.html"
test_endpoint "Training View" "$FRONTEND_URL/training-view.html"
test_endpoint "Evaluation View" "$FRONTEND_URL/evaluation-view.html"

# Test 4: API Integration
echo -e "\n${YELLOW}4. Testing API Integration${NC}"

# Create test dataset
echo -n "Creating test dataset... "
DATASET_RESPONSE=$(curl -s -X POST "$API_URL/datasets" \
  -F "file=@data/contracts/ledgar_finetune_train.json" \
  -F "name=E2E Test Dataset" \
  -F "description=Test dataset for E2E testing")

if echo "$DATASET_RESPONSE" | grep -q "ID"; then
    DATASET_ID=$(echo "$DATASET_RESPONSE" | grep -o '"ID":[0-9]*' | grep -o '[0-9]*')
    echo -e "${GREEN}✓ PASS${NC} (Dataset ID: $DATASET_ID)"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
    DATASET_ID=1  # Fallback
fi

# List datasets
test_endpoint "List Datasets" "$API_URL/datasets"

# Create test job
echo -n "Creating test job... "
JOB_RESPONSE=$(curl -s -X POST "$API_URL/jobs" \
  -H "Content-Type: application/json" \
  -d "{\"dataset_id\": $DATASET_ID, \"configuration\": {\"epochs\": 2}}")

if echo "$JOB_RESPONSE" | grep -q "ID"; then
    JOB_ID=$(echo "$JOB_RESPONSE" | grep -o '"ID":[0-9]*' | grep -o '[0-9]*')
    echo -e "${GREEN}✓ PASS${NC} (Job ID: $JOB_ID)"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
    JOB_ID=1  # Fallback
fi

# Test 5: Real-time Features
echo -e "\n${YELLOW}5. Testing Real-time Features${NC}"

# Test SSE logs (connect for 5 seconds)
echo -n "Testing SSE logs stream... "
timeout 5 curl -N -s "$API_URL/jobs/$JOB_ID/logs" > /tmp/sse_test.log 2>&1 || true

if [ -s /tmp/sse_test.log ]; then
    echo -e "${GREEN}✓ PASS${NC} (SSE connected)"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠ SKIP${NC} (No logs yet)"
fi

# Test 6: Model Operations
echo -e "\n${YELLOW}6. Testing Model Operations${NC}"
test_endpoint "List Models" "$API_URL/models"

# Test 7: Evaluation Operations
echo -e "\n${YELLOW}7. Testing Evaluation Operations${NC}"
test_endpoint "List Evaluations" "$API_URL/evaluations"

# Test 8: Frontend Navigation
echo -e "\n${YELLOW}8. Testing Frontend Navigation${NC}"

# Check if pages have correct content
echo -n "Checking Dashboard content... "
DASHBOARD_CONTENT=$(curl -s "$FRONTEND_URL/index.html")
if echo "$DASHBOARD_CONTENT" | grep -q "Finetune Studio" && \
   echo "$DASHBOARD_CONTENT" | grep -q "Recent Training Jobs"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

echo -n "Checking Upload page content... "
UPLOAD_CONTENT=$(curl -s "$FRONTEND_URL/dataset-upload.html")
if echo "$UPLOAD_CONTENT" | grep -q "Upload Dataset" && \
   echo "$UPLOAD_CONTENT" | grep -q "Drag & drop"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

echo -n "Checking Training page content... "
TRAINING_CONTENT=$(curl -s "$FRONTEND_URL/training-new.html")
if echo "$TRAINING_CONTENT" | grep -q "Start New Training" && \
   echo "$TRAINING_CONTENT" | grep -q "Select Dataset"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

# Test 9: JavaScript Modules
echo -e "\n${YELLOW}9. Testing JavaScript Modules${NC}"

echo -n "Checking API module... "
API_JS=$(curl -s "$FRONTEND_URL/js/api.js")
if echo "$API_JS" | grep -q "class API" && \
   echo "$API_JS" | grep -q "uploadDataset" && \
   echo "$API_JS" | grep -q "streamLogs"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

echo -n "Checking Utils module... "
UTILS_JS=$(curl -s "$FRONTEND_URL/js/utils.js")
if echo "$UTILS_JS" | grep -q "formatTime" && \
   echo "$UTILS_JS" | grep -q "formatBytes" && \
   echo "$UTILS_JS" | grep -q "showToast"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

echo -n "Checking Charts module... "
CHARTS_JS=$(curl -s "$FRONTEND_URL/js/charts.js")
if echo "$CHARTS_JS" | grep -q "createLossChart" && \
   echo "$CHARTS_JS" | grep -q "createComparisonChart"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

# Test 10: CSS Styling
echo -e "\n${YELLOW}10. Testing CSS Styling${NC}"

echo -n "Checking CSS file... "
CSS_CONTENT=$(curl -s "$FRONTEND_URL/css/styles.css")
if echo "$CSS_CONTENT" | grep -q ":root" && \
   echo "$CSS_CONTENT" | grep -q "--primary" && \
   echo "$CSS_CONTENT" | grep -q ".btn"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

# Summary
echo -e "\n=========================================="
echo -e "Test Summary"
echo -e "=========================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo -e "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Some tests failed${NC}"
    exit 1
fi
