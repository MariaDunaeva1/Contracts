#!/bin/bash

# Production Test Script
# Tests all critical functionality after deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:80}"
TIMEOUT=10

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Test functions
test_health_check() {
    print_test "Testing health check endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "$API_URL/api/v1/health")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        # Check if all services are up
        if echo "$body" | grep -q '"status":"healthy"' || echo "$body" | grep -q '"status":"ok"'; then
            print_pass "Health check returned 200 OK"
            
            # Check individual services
            if echo "$body" | grep -q '"database"'; then
                print_info "Database status: $(echo "$body" | grep -o '"database":[^}]*}' || echo 'unknown')"
            fi
            if echo "$body" | grep -q '"storage"'; then
                print_info "Storage status: $(echo "$body" | grep -o '"storage":[^}]*}' || echo 'unknown')"
            fi
        else
            print_fail "Health check returned 200 but status is not healthy"
        fi
    else
        print_fail "Health check failed with status $http_code"
    fi
}

test_metrics_endpoint() {
    print_test "Testing metrics endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "$API_URL/api/v1/metrics")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        if echo "$body" | grep -q "http_requests_total"; then
            print_pass "Metrics endpoint working"
        else
            print_fail "Metrics endpoint returned 200 but no metrics found"
        fi
    else
        print_fail "Metrics endpoint failed with status $http_code"
    fi
}

test_cors_headers() {
    print_test "Testing CORS headers..."
    
    headers=$(curl -s -I -X OPTIONS "$API_URL/api/v1/health")
    
    if echo "$headers" | grep -qi "Access-Control-Allow-Origin"; then
        print_pass "CORS headers present"
    else
        print_fail "CORS headers missing"
    fi
}

test_rate_limiting() {
    print_test "Testing rate limiting..."
    
    # Make multiple rapid requests
    for i in {1..150}; do
        curl -s -o /dev/null -w "%{http_code}\n" "$API_URL/api/v1/health" >> /tmp/rate_test.txt
    done
    
    if grep -q "429" /tmp/rate_test.txt; then
        print_pass "Rate limiting is working (429 Too Many Requests)"
    else
        print_fail "Rate limiting not working (no 429 responses)"
    fi
    
    rm -f /tmp/rate_test.txt
}

test_datasets_endpoint() {
    print_test "Testing datasets list endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "$API_URL/api/v1/datasets")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ]; then
        print_pass "Datasets endpoint working"
    else
        print_fail "Datasets endpoint failed with status $http_code"
    fi
}

test_jobs_endpoint() {
    print_test "Testing jobs list endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "$API_URL/api/v1/jobs")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ]; then
        print_pass "Jobs endpoint working"
    else
        print_fail "Jobs endpoint failed with status $http_code"
    fi
}

test_models_endpoint() {
    print_test "Testing models list endpoint..."
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "$API_URL/api/v1/models")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ]; then
        print_pass "Models endpoint working"
    else
        print_fail "Models endpoint failed with status $http_code"
    fi
}

test_frontend() {
    print_test "Testing frontend accessibility..."
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "$FRONTEND_URL")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        if echo "$body" | grep -q "Finetune Studio" || echo "$body" | grep -q "<!DOCTYPE html>"; then
            print_pass "Frontend is accessible"
        else
            print_fail "Frontend returned 200 but content is unexpected"
        fi
    else
        print_fail "Frontend failed with status $http_code"
    fi
}

test_ssl_certificate() {
    print_test "Testing SSL certificate (if HTTPS)..."
    
    if [[ "$API_URL" == https://* ]]; then
        if curl -s --max-time $TIMEOUT "$API_URL/api/v1/health" > /dev/null 2>&1; then
            print_pass "SSL certificate is valid"
        else
            print_fail "SSL certificate validation failed"
        fi
    else
        print_info "Skipping SSL test (not using HTTPS)"
    fi
}

test_docker_services() {
    print_test "Testing Docker services status..."
    
    if command -v docker &> /dev/null; then
        # Check if running in docker-compose
        if docker compose ps &> /dev/null; then
            services=$(docker compose ps --format json 2>/dev/null || docker compose ps)
            
            if echo "$services" | grep -q "running" || echo "$services" | grep -q "Up"; then
                print_pass "Docker services are running"
            else
                print_fail "Some Docker services are not running"
            fi
        else
            print_info "Not running in docker-compose environment"
        fi
    else
        print_info "Docker not available, skipping service check"
    fi
}

test_response_time() {
    print_test "Testing API response time..."
    
    start_time=$(date +%s%N)
    curl -s --max-time $TIMEOUT "$API_URL/api/v1/health" > /dev/null
    end_time=$(date +%s%N)
    
    duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    if [ $duration -lt 500 ]; then
        print_pass "Response time: ${duration}ms (< 500ms)"
    else
        print_fail "Response time: ${duration}ms (>= 500ms)"
    fi
}

test_log_format() {
    print_test "Testing log format..."
    
    if command -v docker &> /dev/null; then
        if docker compose logs backend 2>/dev/null | tail -n 1 | grep -q "{"; then
            print_pass "Logs are in JSON format"
        else
            print_info "Logs might not be in JSON format (check LOG_FORMAT env var)"
        fi
    else
        print_info "Cannot check log format (Docker not available)"
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "  Production Test Suite"
    echo "=========================================="
    echo ""
    echo "API URL: $API_URL"
    echo "Frontend URL: $FRONTEND_URL"
    echo ""
    
    # Run all tests
    test_health_check
    test_metrics_endpoint
    test_cors_headers
    test_datasets_endpoint
    test_jobs_endpoint
    test_models_endpoint
    test_frontend
    test_response_time
    test_ssl_certificate
    test_docker_services
    test_log_format
    test_rate_limiting
    
    # Summary
    echo ""
    echo "=========================================="
    echo "  Test Summary"
    echo "=========================================="
    echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Run main function
main
