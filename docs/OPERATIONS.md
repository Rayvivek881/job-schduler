# Operations Runbook

Operational guide for maintaining and troubleshooting the Job Scheduler system in production.

## Service Management

### Starting Services

**Kubernetes**:
```bash
# Start all services
kubectl apply -f deploy/k8s/

# Check status
kubectl get pods -l app=job-scheduler
```

**Systemd**:
```bash
# Start services
sudo systemctl start job-scheduler-api
sudo systemctl start job-scheduler-scheduler
sudo systemctl start job-scheduler-consumer@1

# Check status
sudo systemctl status job-scheduler-api
```

**Docker Compose**:
```bash
docker-compose up -d
docker-compose ps
```

### Stopping Services

**Kubernetes**:
```bash
# Graceful shutdown
kubectl delete -f deploy/k8s/
```

**Systemd**:
```bash
sudo systemctl stop job-scheduler-api
sudo systemctl stop job-scheduler-scheduler
sudo systemctl stop job-scheduler-consumer@1
```

**Docker Compose**:
```bash
docker-compose down
```

### Restarting Services

**Kubernetes**:
```bash
kubectl rollout restart deployment job-scheduler-api
kubectl rollout restart deployment job-scheduler-scheduler
kubectl rollout restart deployment job-scheduler-consumer
```

**Systemd**:
```bash
sudo systemctl restart job-scheduler-api
sudo systemctl restart job-scheduler-scheduler
sudo systemctl restart job-scheduler-consumer@1
```

## Scaling

### Adding Consumer Instances

**Kubernetes**:
```bash
# Scale to 5 consumers
kubectl scale deployment job-scheduler-consumer --replicas=5

# Verify
kubectl get pods -l component=consumer
```

**Systemd**:
```bash
# Enable and start additional consumer
sudo systemctl enable job-scheduler-consumer@2
sudo systemctl start job-scheduler-consumer@2

# Check status
sudo systemctl status job-scheduler-consumer@2
```

**Docker Compose**:
```bash
# Scale consumers
docker-compose up -d --scale consumer=5

# Verify
docker-compose ps consumer
```

**Important**: Ensure Kafka topic has enough partitions (at least equal to consumer count).

### Scaling API Server

**Kubernetes**:
```bash
kubectl scale deployment job-scheduler-api --replicas=3
```

**Note**: API server can be scaled horizontally. Use load balancer or ingress to distribute traffic.

### Scheduler Scaling

**DO NOT scale scheduler** - it must remain a singleton (1 replica) to avoid duplicate job scheduling.

## Monitoring

### Check Service Health

```bash
# Health endpoint
curl http://localhost:8000/health

# Expected response:
# {"status":"ok","components":{"database":{"status":"up"},"kafka":{"status":"up","brokers":3}}}
```

### View Logs

**Kubernetes**:
```bash
# All components
kubectl logs -l app=job-scheduler --tail=100 -f

# Specific component
kubectl logs -l component=api-server --tail=100 -f
kubectl logs -l component=scheduler --tail=100 -f
kubectl logs -l component=consumer --tail=100 -f
```

**Systemd**:
```bash
# Follow logs
sudo journalctl -u job-scheduler-api -f
sudo journalctl -u job-scheduler-scheduler -f
sudo journalctl -u job-scheduler-consumer@1 -f

# Last 100 lines
sudo journalctl -u job-scheduler-api -n 100
```

**Docker Compose**:
```bash
docker-compose logs -f api-server
docker-compose logs -f scheduler
docker-compose logs -f consumer
```

### Check Job Status

```sql
-- Job status summary
SELECT status, COUNT(*) as count
FROM jobs
WHERE deleted_at IS NULL
GROUP BY status;

-- Jobs in queue
SELECT COUNT(*) as queued
FROM jobs
WHERE status IN ('in_queue', 'processing')
  AND deleted_at IS NULL;

-- Failed jobs (last 24 hours)
SELECT uuid, job_type, updated_at, job_response->>'runtime_errors' as errors
FROM jobs
WHERE status = 'failed'
  AND updated_at > NOW() - INTERVAL '24 hours'
  AND deleted_at IS NULL
ORDER BY updated_at DESC;
```

### Check Kafka Consumer Lag

```bash
kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group jobs-consumer-group \
  --describe
```

**Healthy**: LAG should be 0 or low (< 1000)

**Action if high**: Scale consumers or investigate processing issues

## Common Operations

### Retry Failed Job

```bash
curl -X PUT http://localhost:8000/jobs/{job-uuid}/retry \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"retry_count": 3}'
```

### Cancel Job

Mark job as deleted (soft delete):

```sql
UPDATE jobs
SET deleted_at = NOW()
WHERE uuid = 'job-uuid';
```

### Check DAG Status

```bash
curl "http://localhost:8000/jobs/dag/status?uuids=job-1&uuids=job-2" \
  -H "X-API-Key: your-api-key"
```

### Clear Old Jobs

```sql
-- Soft delete jobs older than 30 days
UPDATE jobs
SET deleted_at = NOW()
WHERE created_at < NOW() - INTERVAL '30 days'
  AND deleted_at IS NULL;
```

## Troubleshooting

### Issue: Jobs Not Processing

**Symptoms**: Jobs stuck in `open` or `in_queue` status

**Investigation**:
1. Check scheduler is running
2. Check consumers are running
3. Check Kafka connectivity
4. Check consumer logs for errors

**Solutions**:
- Restart scheduler: `kubectl rollout restart deployment job-scheduler-scheduler`
- Restart consumers: `kubectl rollout restart deployment job-scheduler-consumer`
- Check Kafka consumer lag
- Verify topic exists and is accessible

### Issue: High Consumer Lag

**Symptoms**: Consumer lag > 1000 messages

**Investigation**:
```bash
kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --group jobs-consumer-group --describe
```

**Solutions**:
1. Scale consumers: `kubectl scale deployment job-scheduler-consumer --replicas=5`
2. Check job processing time (may be too slow)
3. Verify Kafka topic has enough partitions
4. Check consumer resource usage (CPU/Memory)

### Issue: Database Connection Errors

**Symptoms**: "connection refused" or "too many connections"

**Investigation**:
```sql
-- Check active connections
SELECT count(*) FROM pg_stat_activity WHERE datname = 'job_scheduler';

-- Check connection limit
SHOW max_connections;
```

**Solutions**:
1. Increase connection pool in [`clients/pgsql.go`](../clients/pgsql.go)
2. Check for connection leaks
3. Restart database if needed

### Issue: API High Latency

**Symptoms**: API responses > 1 second

**Investigation**:
1. Check API server logs for slow queries
2. Monitor database query performance
3. Check resource usage (CPU/Memory)

**Solutions**:
1. Scale API server replicas
2. Optimize database queries
3. Add database indexes
4. Review rate limiting settings

### Issue: Jobs Failing Immediately

**Symptoms**: Jobs marked as `failed` right after creation

**Investigation**:
```sql
-- Check error messages
SELECT uuid, job_type, job_response->>'runtime_errors' as errors
FROM jobs
WHERE status = 'failed'
ORDER BY updated_at DESC
LIMIT 10;
```

**Solutions**:
1. Check processor implementation for the job type
2. Verify job data format is correct
3. Check dependencies (if job requires external services)
4. Review job execution timeout settings

## Backup and Restore

### Database Backup

```bash
# Full backup
pg_dump -U postgres job_scheduler > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup with compression
pg_dump -U postgres job_scheduler | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

### Database Restore

```bash
# Restore from backup
psql -U postgres job_scheduler < backup_20240115_120000.sql

# Or from compressed backup
gunzip -c backup_20240115_120000.sql.gz | psql -U postgres job_scheduler
```

### Backup Strategy

**Recommended**:
- Daily full backups
- Keep last 7 daily backups
- Weekly backups for last 4 weeks
- Monthly backups for last 12 months

## Maintenance Windows

### Planned Maintenance

1. **Notify users** of maintenance window
2. **Scale down services** (optional, for zero-downtime):
   ```bash
   kubectl scale deployment job-scheduler-api --replicas=0
   kubectl scale deployment job-scheduler-consumer --replicas=0
   ```

3. **Perform maintenance** (schema updates, configuration changes)

4. **Verify changes**:
   ```bash
   bash deploy/verify.sh
   ```

5. **Scale services back up**

6. **Monitor** for issues

### Zero-Downtime Updates

**API Server**:
```bash
# Rolling update (Kubernetes handles this automatically)
kubectl set image deployment/job-scheduler-api api-server=job-scheduler:v1.1.0
kubectl rollout status deployment/job-scheduler-api
```

**Consumers**:
```bash
# Update consumers (Kafka handles rebalancing)
kubectl set image deployment/job-scheduler-consumer consumer=job-scheduler:v1.1.0
kubectl rollout status deployment/job-scheduler-consumer
```

**Scheduler**:
```bash
# Scheduler must be updated carefully (brief downtime acceptable)
kubectl set image deployment/job-scheduler-scheduler scheduler=job-scheduler:v1.1.0
```

## Performance Tuning

### Database Performance

**Indexes**: Ensure indexes are created (see `deploy/sql/init_schema.sql`)

**Connection Pool**: Adjust in [`clients/pgsql.go`](../clients/pgsql.go):
- `MaxOpenConns`: 40 (current)
- `MaxIdleConns`: 20 (current)

### Kafka Performance

**Partitions**: Ensure topic has enough partitions for consumer count

**Consumer Settings**: Adjust consumer count based on load

### Application Performance

**Resource Limits**: Adjust in deployment manifests:
- API Server: 512Mi memory, 500m CPU
- Consumers: 512Mi memory, 500m CPU
- Scheduler: 256Mi memory, 200m CPU

## Security

### Rotate API Key

1. Update secret:
```bash
kubectl create secret generic job-scheduler-secrets \
  --from-literal=API_KEY='new-api-key' \
  --dry-run=client -o yaml | kubectl apply -f -
```

2. Restart API server:
```bash
kubectl rollout restart deployment/job-scheduler-api
```

### Update Database Password

1. Update password in database
2. Update secret:
```bash
kubectl create secret generic job-scheduler-secrets \
  --from-literal=PG_DB_PASSWORD='new-password' \
  --dry-run=client -o yaml | kubectl apply -f -
```

3. Restart all services:
```bash
kubectl rollout restart deployment -l app=job-scheduler
```

## Emergency Procedures

### Complete System Failure

1. **Assess damage**: Check logs, database, Kafka
2. **Restore from backup** if needed
3. **Restart services** in order:
   - Database
   - Kafka
   - Scheduler
   - API Server
   - Consumers

### Data Corruption

1. **Stop all services**
2. **Restore database from backup**
3. **Verify data integrity**
4. **Restart services**

### Kafka Topic Corruption

1. **Stop consumers**
2. **Delete and recreate topic** (if acceptable to lose in-flight jobs)
3. **Restart consumers**

## Contact and Escalation

- **On-Call Engineer**: [Contact Information]
- **Database Admin**: [Contact Information]
- **Kafka Admin**: [Contact Information]

## References

- [Monitoring Guide](MONITORING.md)
- [Configuration Guide](CONFIGURATION.md)
- [API Documentation](API.md)
