#!/bin/bash
# Kafka Topic Creation Script
# Automates creation of Kafka topic for Job Scheduler with recommended settings
#
# Usage: bash deploy/kafka/create-topic.sh [options]
#   --bootstrap-server <server>  Kafka bootstrap server (required)
#   --topic <topic-name>         Topic name (default: jobs-topic)
#   --partitions <count>         Number of partitions (default: 6)
#   --replication-factor <count> Replication factor (default: 3 for production, 1 for dev)
#   --environment <env>          Environment: development, staging, production (default: development)
#   --dry-run                    Show command without executing
#
# Examples:
#   # Development (single broker)
#   bash deploy/kafka/create-topic.sh --bootstrap-server localhost:9092 --environment development
#
#   # Production
#   bash deploy/kafka/create-topic.sh --bootstrap-server kafka1:9092,kafka2:9092,kafka3:9092 --environment production

set -e

# Default values
BOOTSTRAP_SERVER=""
TOPIC_NAME="jobs-topic"
PARTITIONS=6
REPLICATION_FACTOR=1
ENVIRONMENT="development"
DRY_RUN=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --bootstrap-server)
            BOOTSTRAP_SERVER="$2"
            shift 2
            ;;
        --topic)
            TOPIC_NAME="$2"
            shift 2
            ;;
        --partitions)
            PARTITIONS="$2"
            shift 2
            ;;
        --replication-factor)
            REPLICATION_FACTOR="$2"
            shift 2
            ;;
        --environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --bootstrap-server <server>  Kafka bootstrap server (required)"
            echo "  --topic <topic-name>         Topic name (default: jobs-topic)"
            echo "  --partitions <count>         Number of partitions (default: 6)"
            echo "  --replication-factor <count> Replication factor (default: 1 for dev, 3 for prod)"
            echo "  --environment <env>          Environment: development, staging, production"
            echo "  --dry-run                    Show command without executing"
            echo ""
            echo "Examples:"
            echo "  # Development"
            echo "  $0 --bootstrap-server localhost:9092 --environment development"
            echo ""
            echo "  # Production"
            echo "  $0 --bootstrap-server kafka1:9092,kafka2:9092,kafka3:9092 --environment production"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Validate required arguments
if [ -z "$BOOTSTRAP_SERVER" ]; then
    echo -e "${RED}ERROR: --bootstrap-server is required${NC}"
    echo "Usage: $0 --bootstrap-server <server> [options]"
    exit 1
fi

# Set defaults based on environment
case "$ENVIRONMENT" in
    production|prod)
        if [ "$REPLICATION_FACTOR" -eq 1 ]; then
            REPLICATION_FACTOR=3
        fi
        RETENTION_MS=604800000  # 7 days
        COMPRESSION_TYPE="snappy"
        ;;
    staging|stage)
        if [ "$REPLICATION_FACTOR" -eq 1 ]; then
            REPLICATION_FACTOR=2
        fi
        RETENTION_MS=259200000  # 3 days
        COMPRESSION_TYPE="snappy"
        ;;
    development|dev)
        REPLICATION_FACTOR=1
        RETENTION_MS=86400000  # 1 day
        COMPRESSION_TYPE="uncompressed"
        ;;
    *)
        echo -e "${YELLOW}WARNING: Unknown environment '$ENVIRONMENT', using development defaults${NC}"
        ENVIRONMENT="development"
        REPLICATION_FACTOR=1
        RETENTION_MS=86400000
        COMPRESSION_TYPE="uncompressed"
        ;;
esac

echo "=========================================="
echo "Kafka Topic Creation"
echo "=========================================="
echo "Bootstrap Server: $BOOTSTRAP_SERVER"
echo "Topic Name: $TOPIC_NAME"
echo "Partitions: $PARTITIONS"
echo "Replication Factor: $REPLICATION_FACTOR"
echo "Environment: $ENVIRONMENT"
echo "Retention: ${RETENTION_MS}ms ($(($RETENTION_MS / 86400000)) days)"
echo "Compression: $COMPRESSION_TYPE"
echo ""

# Check if topic already exists
if kafka-topics.sh --bootstrap-server "$BOOTSTRAP_SERVER" --list 2>/dev/null | grep -q "^${TOPIC_NAME}$"; then
    echo -e "${YELLOW}WARNING: Topic '$TOPIC_NAME' already exists${NC}"
    echo ""
    echo "Current topic configuration:"
    kafka-topics.sh --bootstrap-server "$BOOTSTRAP_SERVER" --describe --topic "$TOPIC_NAME" 2>/dev/null || true
    echo ""
    read -p "Do you want to recreate it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Exiting without changes."
        exit 0
    fi
    
    if [ "$DRY_RUN" = false ]; then
        echo "Deleting existing topic..."
        kafka-topics.sh --bootstrap-server "$BOOTSTRAP_SERVER" --delete --topic "$TOPIC_NAME" 2>/dev/null || true
        echo "Waiting for topic deletion to complete..."
        sleep 5
    fi
fi

# Build kafka-topics command
CMD="kafka-topics.sh --create \
  --bootstrap-server $BOOTSTRAP_SERVER \
  --topic $TOPIC_NAME \
  --partitions $PARTITIONS \
  --replication-factor $REPLICATION_FACTOR \
  --config retention.ms=$RETENTION_MS \
  --config compression.type=$COMPRESSION_TYPE \
  --config cleanup.policy=delete"

echo "Command to execute:"
echo "$CMD"
echo ""

if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}DRY RUN: Not executing command${NC}"
    exit 0
fi

# Execute command
echo "Creating topic..."
if eval "$CMD"; then
    echo ""
    echo -e "${GREEN}✓ Topic created successfully${NC}"
    echo ""
    echo "Verifying topic configuration:"
    kafka-topics.sh --bootstrap-server "$BOOTSTRAP_SERVER" --describe --topic "$TOPIC_NAME"
    echo ""
    echo -e "${GREEN}Topic '$TOPIC_NAME' is ready to use${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Set JOBS_TOPIC=$TOPIC_NAME in your environment"
    echo "2. Set CONSUMER_GROUP (e.g., jobs-consumer-group) in your environment"
    echo "3. Start your consumers"
else
    echo ""
    echo -e "${RED}✗ Failed to create topic${NC}"
    exit 1
fi
