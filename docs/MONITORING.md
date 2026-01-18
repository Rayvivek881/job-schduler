# Monitoring and Observability Guide

Complete guide for monitoring the Job Scheduler system in production.

## Health Checks

### Health Endpoint

The API server exposes a `/health` endpoint that checks component status:

```bash
curl http://localhost:8000/health
```

**Response** (200 OK):
```json
{
  "status": "ok",
  "timestamp": {
    "current": "2024-01-15T10:00:00Z"
  },
  "components": {
    "database": {
      "status": "up"
    },
    "kafka": {
      "status": "up",
      "brokers": 3
    }
  }
}
```

**Response** (503 Service Unavailable):
```json
{
  "status": "degraded",
  "components": {
    "database": {
      "status": "down",
      "error": "connection refused"
    }
  }
}
```

### Component Status Values

- `ok`: All components healthy
- `degraded`: Some components unhealthy but system can partially function
- `down`: Critical components unavailable

## Logging

### Log Structure

The system uses structured logging via `zerolog`. All logs are in JSON format:

```json
{
  "level": "info",
  "time": "2024-01-15T10:00:00Z",
  "job_uuid": "job-123",
  "job_type": "etl",
  "message": "Job processed successfully"
}
```

### Log Levels

- **ERROR**: Critical errors requiring attention
- **WARN**: Warning conditions
- **INFO**: Informational messages (default)
- **DEBUG**: Detailed debugging information

### Setting Log Level

Set log level via environment or code configuration:

```go
// In production, log level is controlled by APP_ENV
// Development: DEBUG
// Production: INFO (recommended)
```

### Log Aggregation

#### CloudWatch (AWS)

Configure CloudWatch Logs for ECS/Fargate deployments:

```yaml
# In ECS task definition
logConfiguration:
  logDriver: awslogs
  options:
    awslogs-group: /ecs/job-scheduler
    awslogs-region: us-east-1
    awslogs-stream-prefix: api-server
```

#### ELK Stack

Ship logs to Elasticsearch/Logstash:

```bash
# Using Filebeat
filebeat:
  inputs:
    - type: container
      paths:
        - /var/lib/docker/containers/*/*.log
```

#### Kubernetes

View logs in Kubernetes:

```bash
# All components
kubectl logs -l app=job-scheduler --tail=100

# Specific component
kubectl logs -l component=api-server --tail=100 -f
kubectl logs -l component=scheduler --tail=100 -f
kubectl logs -l component=consumer --tail=100 -f
```

## Metrics (Recommended)

### Prometheus Metrics (Future Enhancement)

Recommended metrics to expose:

#### API Server Metrics

```go
// Request rate
http_requests_total{method, endpoint, status}

// Request latency
http_request_duration_seconds{method, endpoint}

// Active connections
http_connections_active
```

#### Job Metrics

```go
// Job status counts
job_status_total{status}

// Job processing duration
job_processing_duration_seconds{job_type}

// Jobs processed
jobs_processed_total{job_type, status}

// Job queue depth
job_queue_depth
```

#### Kafka Metrics

```go
// Consumer lag
kafka_consumer_lag{partition}

// Consumer lag time
kafka_consumer_lag_time_seconds{partition}

// Messages consumed
kafka_messages_consumed_total{partition}
```

#### Database Metrics

```go
// Database connections
db_connections_active
db_connections_idle

// Query duration
db_query_duration_seconds{query_type}

// Query errors
db_query_errors_total{query_type}
```

### Current Monitoring Approach

Currently, monitoring can be done via:

1. **Health Checks**: `/health` endpoint for liveness/readiness
2. **Application Logs**: Structured logging via zerolog
3. **Database Queries**: Direct PostgreSQL queries for job status
4. **Kafka Consumer Groups**: `kafka-consumer-groups.sh` for consumer lag

## Key Metrics to Monitor

### API Server

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| Request rate | Requests per second | > 1000 req/s sustained |
| Error rate | 5xx errors / total requests | > 1% |
| Response time (p95) | 95th percentile latency | > 1s |
| Health check failures | Failed `/health` checks | > 3 consecutive |

### Jobs

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| Job completion rate | Jobs completed / hour | < expected rate |
| Job failure rate | Failed jobs / total | > 5% |
| Job processing time | Average job duration | > timeout threshold |
| Queue depth | Jobs in `open` status | > 10,000 |
| Consumer lag | Unprocessed messages | > 1000 |

### Infrastructure

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| Database connections | Active DB connections | > 80% of max (40) |
| Kafka consumer lag | Partition lag | > 1000 messages |
| Memory usage | Container memory | > 80% of limit |
| CPU usage | Container CPU | > 80% sustained |

## Alerting Recommendations

### Critical Alerts

1. **API Server Down**: Health check failing for > 5 minutes
2. **Database Unavailable**: Cannot connect to PostgreSQL
3. **Kafka Unavailable**: Cannot connect to Kafka brokers
4. **High Consumer Lag**: Consumer lag > 10,000 messages
5. **Job Failure Spike**: Failure rate > 10% in last hour

### Warning Alerts

1. **Degraded Health**: Health endpoint returns `degraded`
2. **High Error Rate**: API error rate > 1%
3. **Slow Job Processing**: Average job time > threshold
4. **Resource Exhaustion**: CPU/Memory > 80%

## Monitoring Queries

### Database Queries

#### Job Status Summary

```sql
SELECT 
    status, 
    COUNT(*) as count,
    AVG(EXTRACT(EPOCH FROM (updated_at - created_at))) as avg_duration_seconds
FROM jobs
WHERE deleted_at IS NULL
GROUP BY status;
```

#### Failed Jobs (Last 24 Hours)

```sql
SELECT 
    uuid,
    job_type,
    job_title,
    retry_count,
    updated_at,
    job_response->>'runtime_errors' as errors
FROM jobs
WHERE status = 'failed'
  AND updated_at > NOW() - INTERVAL '24 hours'
  AND deleted_at IS NULL
ORDER BY updated_at DESC;
```

#### Jobs Pending (Degree = 0, Ready to Run)

```sql
SELECT COUNT(*) as pending_jobs
FROM jobs
WHERE degree = 0
  AND status = 'open'
  AND run_after <= NOW()
  AND deleted_at IS NULL;
```

#### Consumer Lag Estimation

```sql
-- Jobs in queue but not processing
SELECT COUNT(*) as jobs_in_queue
FROM jobs
WHERE status IN ('in_queue', 'processing')
  AND deleted_at IS NULL;
```

### Kafka Consumer Group Status

```bash
# Check consumer lag
kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group jobs-consumer-group \
  --describe

# Monitor lag in real-time
watch -n 5 'kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group jobs-consumer-group \
  --describe | grep LAG'
```

## Dashboards

### Recommended Dashboard Panels

1. **System Health**
   - API server health status
   - Database connectivity
   - Kafka connectivity

2. **Job Metrics**
   - Jobs by status (pie chart)
   - Job completion rate over time
   - Average job processing time
   - Failed jobs count

3. **Performance**
   - API request rate
   - API response time (p50, p95, p99)
   - Error rate

4. **Queue Status**
   - Queue depth (jobs waiting)
   - Consumer lag
   - Jobs processed per hour

5. **Resource Usage**
   - CPU usage (per component)
   - Memory usage (per component)
   - Database connection pool usage

## Troubleshooting

### High Consumer Lag

**Symptoms**: Jobs queued but not processing

**Investigation**:
1. Check consumer logs for errors
2. Verify Kafka connectivity
3. Check consumer resource usage (CPU/Memory)
4. Verify consumer instances are running

**Solutions**:
- Scale up consumers: `kubectl scale deployment job-scheduler-consumer --replicas=5`
- Check Kafka topic partitions (need >= consumer count)
- Review job processing time (may be too slow)

### Database Connection Pool Exhausted

**Symptoms**: Connection errors in logs

**Investigation**:
```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = 'job_scheduler';
```

**Solutions**:
- Increase `MaxOpenConns` in [`clients/pgsql.go`](../clients/pgsql.go)
- Review connection pool settings
- Check for connection leaks

### API High Latency

**Investigation**:
1. Check API server logs for slow queries
2. Monitor database query performance
3. Check resource usage

**Solutions**:
- Scale API server replicas
- Optimize database queries
- Add database indexes
- Review rate limiting settings

## References

- Health endpoint: [`cmd/server.go`](../cmd/server.go)
- Logging: Uses `zerolog` throughout codebase
- Database connection: [`clients/pgsql.go`](../clients/pgsql.go)
- Kafka client: [`clients/kafka.go`](../clients/kafka.go)
