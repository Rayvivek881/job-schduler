#!/bin/bash
# Deployment Verification Script
# Tests all components of the Job Scheduler system
#
# Usage: bash deploy/verify.sh [API_URL]
#   API_URL: Base URL for API server (default: http://localhost:8000)

set -e

API_URL="${1:-http://localhost:8000}"
API_KEY="${API_KEY:-your-api-key}"

echo "=========================================="
echo "Job Scheduler Deployment Verification"
echo "=========================================="
echo "API URL: $API_URL"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to test endpoint
test_endpoint() {
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    local expected_status="${5:-200}"
    
    echo -n "Testing $name... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "X-API-Key: $API_KEY" \
            -d "$data" \
            "$url" 2>/dev/null || echo -e "\n000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "X-API-Key: $API_KEY" \
            "$url" 2>/dev/null || echo -e "\n000")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}PASS${NC} (HTTP $http_code)"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}FAIL${NC} (HTTP $http_code, expected $expected_status)"
        echo "  Response: $body"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test 1: Health Check (no auth required)
echo "1. Health Check"
echo "----------------------------------------"
test_endpoint "Health endpoint" "GET" "$API_URL/health"

# Parse health response
health_status=$(curl -s "$API_URL/health" | grep -o '"status":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
echo "   Health status: $health_status"

if [ "$health_status" != "ok" ] && [ "$health_status" != "degraded" ]; then
    echo -e "   ${YELLOW}WARNING: Health status is not 'ok'${NC}"
fi

# Test 2: Database Connectivity (from health response)
db_status=$(curl -s "$API_URL/health" | grep -o '"database":{"status":"[^"]*"' | cut -d'"' -f6 || echo "unknown")
echo "   Database status: $db_status"

if [ "$db_status" != "up" ]; then
    echo -e "   ${RED}ERROR: Database is not up${NC}"
fi

# Test 3: Kafka Connectivity (from health response)
kafka_status=$(curl -s "$API_URL/health" | grep -o '"kafka":{"status":"[^"]*"' | cut -d'"' -f6 || echo "unknown")
echo "   Kafka status: $kafka_status"

if [ "$kafka_status" != "up" ]; then
    echo -e "   ${YELLOW}WARNING: Kafka is not up${NC}"
fi

echo ""

# Test 4: Create Test DAG
echo "2. DAG Creation"
echo "----------------------------------------"
test_uuid=$(date +%s)
dag_payload="[
  {
    \"uuid\": \"test-job-a-$test_uuid\",
    \"job_title\": \"Test Job A\",
    \"job_type\": \"test\",
    \"data\": {\"test\": true},
    \"retry_count\": 3,
    \"retry_interval\": 5,
    \"edges\": [\"test-job-b-$test_uuid\"]
  },
  {
    \"uuid\": \"test-job-b-$test_uuid\",
    \"job_title\": \"Test Job B\",
    \"job_type\": \"test\",
    \"data\": {\"test\": true},
    \"retry_count\": 3,
    \"retry_interval\": 5,
    \"edges\": []
  }
]"

test_endpoint "Create DAG" "POST" "$API_URL/jobs/bulk-insert/complete-graph" "$dag_payload" "200"

JOB_A_UUID="test-job-a-$test_uuid"
JOB_B_UUID="test-job-b-$test_uuid"

echo ""

# Test 5: Get Job by UUID
echo "3. Job Retrieval"
echo "----------------------------------------"
test_endpoint "Get job A" "GET" "$API_URL/jobs/$JOB_A_UUID"
test_endpoint "Get job B" "GET" "$API_URL/jobs/$JOB_B_UUID"

echo ""

# Test 6: List Jobs with Filters
echo "4. Job Filtering"
echo "----------------------------------------"
test_endpoint "List jobs by UUID" "GET" "$API_URL/jobs/?uuid=$JOB_A_UUID&uuid=$JOB_B_UUID"
test_endpoint "List jobs by status" "GET" "$API_URL/jobs/?status=open"

echo ""

# Test 7: DAG Status
echo "5. DAG Status"
echo "----------------------------------------"
test_endpoint "DAG status" "GET" "$API_URL/jobs/dag/status?uuids=$JOB_A_UUID&uuids=$JOB_B_UUID"
test_endpoint "DAG progress" "GET" "$API_URL/jobs/dag/progress?uuids=$JOB_A_UUID&uuids=$JOB_B_UUID"

echo ""

# Test 8: Retry Job (if any failed jobs exist)
echo "6. Job Retry"
echo "----------------------------------------"
# Create a test job that we can retry
retry_uuid="test-retry-$test_uuid"
retry_payload="[
  {
    \"uuid\": \"$retry_uuid\",
    \"job_title\": \"Test Retry Job\",
    \"job_type\": \"test\",
    \"data\": {\"test\": true},
    \"retry_count\": 1,
    \"retry_interval\": 1,
    \"edges\": []
  }
]"

test_endpoint "Create retry test job" "POST" "$API_URL/jobs/bulk-insert/complete-graph" "$retry_payload" "200"

# Note: Actual retry test would require job to fail first, which is complex in automated test
# This test verifies the endpoint exists
test_endpoint "Retry endpoint accessible" "PUT" "$API_URL/jobs/$retry_uuid/retry" "{\"retry_count\": 2}" "200"

echo ""

# Summary
echo "=========================================="
echo "Verification Summary"
echo "=========================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Please review the output above.${NC}"
    exit 1
fi
