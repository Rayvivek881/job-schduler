# Kafka Topic Configuration

This document describes the Kafka topic setup required for the Job Scheduler system.

## Topic Configuration

### Jobs Topic

**Topic Name**: Configurable via `JOBS_TOPIC` environment variable (default: `jobs-topic`)

#### Recommended Production Settings

```bash
# Create topic with proper configuration
kafka-topics.sh --create \
  --bootstrap-server localhost:9092 \
  --topic jobs-topic \
  --partitions 6 \
  --replication-factor 3 \
  --config retention.ms=604800000 \
  --config compression.type=snappy
```

#### Configuration Parameters

| Parameter | Recommended Value | Description |
|-----------|------------------|-------------|
| `partitions` | 3-6 | Number of partitions for parallel processing. More partitions = more parallel consumers |
| `replication-factor` | 3 (production), 1 (development) | Number of replicas for fault tolerance |
| `retention.ms` | 604800000 (7 days) | How long messages are retained |
| `compression.type` | snappy or lz4 | Compression algorithm for messages |
| `cleanup.policy` | delete | Delete old messages when retention expires |

#### Partition Count Considerations

- **Minimum**: 3 partitions (for basic parallelism)
- **Recommended**: 6 partitions (supports up to 6 concurrent consumers efficiently)
- **Maximum**: Match number of consumers (each partition consumed by one consumer in a consumer group)

**Rule of thumb**: Number of partitions = expected number of consumers

## Consumer Group

**Consumer Group Name**: Configurable via `CONSUMER_GROUP` environment variable (default: `jobs-consumer-group`)

### Behavior

- All consumer instances should use the **same consumer group name**
- Kafka automatically distributes partitions among consumers in the group
- Each partition is consumed by exactly one consumer in the group
- Adding more consumers (replicas) automatically rebalances partitions

### Example

With 6 partitions and consumer group `jobs-consumer-group`:
- 1 consumer → handles all 6 partitions
- 2 consumers → each handles 3 partitions
- 3 consumers → each handles 2 partitions
- 6 consumers → each handles 1 partition
- 10 consumers → 6 handle 1 partition, 4 idle (no partitions available)

## Topic Management Commands

### Create Topic (Automated - Recommended)

Use the automated script for easier setup with environment-specific defaults:

```bash
# Development (single broker)
bash deploy/kafka/create-topic.sh \
  --bootstrap-server localhost:9092 \
  --environment development

# Staging
bash deploy/kafka/create-topic.sh \
  --bootstrap-server kafka-staging:9092 \
  --environment staging \
  --partitions 6

# Production (multi-broker)
bash deploy/kafka/create-topic.sh \
  --bootstrap-server kafka1:9092,kafka2:9092,kafka3:9092 \
  --environment production \
  --partitions 6 \
  --replication-factor 3

# Dry run (show command without executing)
bash deploy/kafka/create-topic.sh \
  --bootstrap-server localhost:9092 \
  --environment development \
  --dry-run
```

The script automatically sets:
- **Development**: replication-factor=1, retention=1 day, compression=uncompressed
- **Staging**: replication-factor=2, retention=3 days, compression=snappy
- **Production**: replication-factor=3, retention=7 days, compression=snappy

### Create Topic (Manual)

```bash
# Basic creation
kafka-topics.sh --create \
  --bootstrap-server <broker>:9092 \
  --topic jobs-topic \
  --partitions 6 \
  --replication-factor 3

# With custom retention
kafka-topics.sh --create \
  --bootstrap-server <broker>:9092 \
  --topic jobs-topic \
  --partitions 6 \
  --replication-factor 3 \
  --config retention.ms=604800000 \
  --config compression.type=snappy \
  --config cleanup.policy=delete
```

### List Topics

```bash
kafka-topics.sh --list --bootstrap-server <broker>:9092
```

### Describe Topic

```bash
kafka-topics.sh --describe \
  --bootstrap-server <broker>:9092 \
  --topic jobs-topic
```

### Check Consumer Group Status

```bash
# List consumer groups
kafka-consumer-groups.sh --bootstrap-server <broker>:9092 --list

# Describe consumer group (shows lag, partition assignment)
kafka-consumer-groups.sh --bootstrap-server <broker>:9092 \
  --group jobs-consumer-group \
  --describe
```

### Delete Topic (Development Only)

```bash
# WARNING: Only use in development/testing
kafka-topics.sh --delete \
  --bootstrap-server <broker>:9092 \
  --topic jobs-topic
```

## Monitoring

### Consumer Lag

Monitor consumer lag to ensure consumers are keeping up:

```bash
kafka-consumer-groups.sh --bootstrap-server <broker>:9092 \
  --group jobs-consumer-group \
  --describe | grep LAG
```

**Healthy state**: LAG should be 0 or low (under 1000)

### Topic Metrics

Monitor these metrics in production:
- **Messages per second**: Production rate
- **Bytes per second**: Throughput
- **Consumer lag**: Processing backlog
- **Partition distribution**: Ensure even distribution

## Security (Production)

### SASL Authentication

If using SASL authentication, configure in environment variables:

```env
KAFKA_USERNAME=your-username
KAFKA_PASSWORD=your-password
KAFKA_SASL_MECHANISM=SCRAM-SHA-256  # or PLAIN
```

### TLS/SSL

TLS is automatically enabled when SASL is configured (see [`clients/kafka.go`](../clients/kafka.go)).

## Environment Variables

Reference configuration in [`conf/viper.go`](../conf/viper.go):

```env
# Kafka Brokers (comma-separated)
KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092

# Kafka Version
KAFKA_VERSION=3.6.0

# Topic and Consumer Group
JOBS_TOPIC=jobs-topic-prod
CONSUMER_GROUP=jobs-consumer-group-prod

# Optional: SASL Authentication
KAFKA_USERNAME=kafka-user
KAFKA_PASSWORD=kafka-password
KAFKA_SASL_MECHANISM=SCRAM-SHA-256
```

## Troubleshooting

### Topic Not Found

**Error**: `topic does not exist`

**Solution**: Create the topic using commands above

### Consumer Not Receiving Messages

**Possible causes**:
1. Consumer group is not subscribed to the topic
2. No messages being produced to the topic
3. Consumer is not part of the consumer group

**Check**:
```bash
# Verify consumer group status
kafka-consumer-groups.sh --bootstrap-server <broker>:9092 \
  --group jobs-consumer-group \
  --describe

# Check if messages are being produced
kafka-console-consumer.sh --bootstrap-server <broker>:9092 \
  --topic jobs-topic \
  --from-beginning
```

### High Consumer Lag

**Symptoms**: Jobs are queued but not processing

**Solutions**:
1. Add more consumer instances (horizontal scaling)
2. Increase partitions (requires topic recreation)
3. Check consumer performance (CPU, memory, I/O)

## References

- Kafka Topic Documentation: https://kafka.apache.org/documentation/#topicconfigs
- Consumer Groups: https://kafka.apache.org/documentation/#consumerconfigs
- Kafka client implementation: [`clients/kafka.go`](../clients/kafka.go)
