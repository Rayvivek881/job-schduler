#!/bin/bash
# Environment Variable Validation Script
# Validates .env file completeness against required variables from conf/viper.go
#
# Usage: bash deploy/validate-env.sh [.env-file-path]
#   .env-file-path: Path to .env file (default: .env in current directory)

set -e

ENV_FILE="${1:-.env}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Environment Variable Validation"
echo "=========================================="
echo "Validating: $ENV_FILE"
echo ""

if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}ERROR: Environment file not found: $ENV_FILE${NC}"
    echo "Usage: bash deploy/validate-env.sh [.env-file-path]"
    echo ""
    echo "Create .env file from template:"
    echo "  cp deploy/env.example .env"
    exit 1
fi

# Counter
VALID_COUNT=0
MISSING_COUNT=0
WARNING_COUNT=0

# Function to check if variable is set
check_var() {
    local var_name="$1"
    local required="${2:-true}"
    local description="${3:-}"
    
    # Check if variable is set (not empty)
    if grep -q "^${var_name}=" "$ENV_FILE" && ! grep "^${var_name}=\s*$" "$ENV_FILE" > /dev/null && ! grep "^${var_name}=$" "$ENV_FILE" > /dev/null; then
        value=$(grep "^${var_name}=" "$ENV_FILE" | cut -d'=' -f2- | tr -d '"' | tr -d "'")
        if [ -n "$value" ] && [ "$value" != "" ]; then
            if [ "$required" == "true" ]; then
                echo -e "  ${GREEN}✓${NC} ${var_name} (required)"
                ((VALID_COUNT++))
            else
                echo -e "  ${GREEN}✓${NC} ${var_name} (optional): ${value}"
                ((VALID_COUNT++))
            fi
            return 0
        fi
    fi
    
    if [ "$required" == "true" ]; then
        echo -e "  ${RED}✗${NC} ${var_name} (required) - MISSING"
        if [ -n "$description" ]; then
            echo "      ${description}"
        fi
        ((MISSING_COUNT++))
        return 1
    else
        echo -e "  ${YELLOW}⚠${NC} ${var_name} (optional) - not set"
        if [ -n "$description" ]; then
            echo "      ${description}"
        fi
        ((WARNING_COUNT++))
        return 2
    fi
}

echo "Required Variables:"
echo "----------------------------------------"

# Required Application Variables
check_var "API_KEY" "true" "API authentication key"
check_var "PG_DB_HOST" "true" "PostgreSQL host address"
check_var "PG_DB_DATABASE" "true" "Database name"
check_var "PG_DB_USERNAME" "true" "Database username"
check_var "PG_DB_PASSWORD" "true" "Database password"
check_var "KAFKA_BROKERS" "true" "Kafka broker addresses (comma-separated)"
check_var "KAFKA_VERSION" "true" "Kafka version (e.g., 3.6.0)"
check_var "JOBS_TOPIC" "true" "Kafka topic name for jobs"
check_var "CONSUMER_GROUP" "true" "Kafka consumer group name"

echo ""
echo "Optional Variables (with defaults):"
echo "----------------------------------------"

# Optional Application Variables
check_var "APP_ENV" "false" "Environment: development, staging, production (default: development)"
check_var "MAX_REQUESTS_PER_MINUTE" "false" "API rate limit (default: 1000)"
check_var "PG_DB_PORT" "false" "PostgreSQL port (default: 5432)"
check_var "PG_DB_DEBUG" "false" "Enable SQL query logging (default: false)"
check_var "PG_DB_SSL" "false" "Enable SSL/TLS connection (default: false)"
check_var "KAFKA_USERNAME" "false" "Kafka SASL username (optional)"
check_var "KAFKA_PASSWORD" "false" "Kafka SASL password (optional)"
check_var "KAFKA_SASL_MECHANISM" "false" "SASL mechanism: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512"
check_var "TICKER_INTERVAL" "false" "Scheduler tick interval in minutes (default: 1)"
check_var "JOB_EXECUTION_TIMEOUT" "false" "Job execution timeout in minutes (default: 30)"
check_var "JOB_IN_QUEUE_SIZE" "false" "Maximum jobs in queue (default: 1000)"
check_var "PARALLEL_JOBS" "false" "Maximum parallel jobs (default: 10)"

echo ""
echo "Kafka SASL Configuration:"
echo "----------------------------------------"
echo "Note: If KAFKA_USERNAME is set, KAFKA_PASSWORD and KAFKA_SASL_MECHANISM are required"

if grep -q "^KAFKA_USERNAME=" "$ENV_FILE" && [ -n "$(grep "^KAFKA_USERNAME=" "$ENV_FILE" | cut -d'=' -f2- | tr -d '"' | tr -d "'" | tr -d ' ')" ]; then
    check_var "KAFKA_PASSWORD" "true" "Required when KAFKA_USERNAME is set"
    check_var "KAFKA_SASL_MECHANISM" "true" "Required when KAFKA_USERNAME is set"
fi

echo ""
echo "=========================================="
echo "Validation Summary"
echo "=========================================="
echo -e "Valid: ${GREEN}${VALID_COUNT}${NC}"
echo -e "Missing (required): ${RED}${MISSING_COUNT}${NC}"
echo -e "Warnings (optional): ${YELLOW}${WARNING_COUNT}${NC}"
echo ""

if [ $MISSING_COUNT -gt 0 ]; then
    echo -e "${RED}Validation FAILED: ${MISSING_COUNT} required variable(s) missing${NC}"
    echo ""
    echo "Please set the missing variables in $ENV_FILE"
    echo "See deploy/env.example for reference"
    exit 1
else
    echo -e "${GREEN}Validation PASSED: All required variables are set${NC}"
    if [ $WARNING_COUNT -gt 0 ]; then
        echo -e "${YELLOW}Note: ${WARNING_COUNT} optional variable(s) not set (will use defaults)${NC}"
    fi
    exit 0
fi
