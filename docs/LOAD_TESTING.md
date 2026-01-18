# Load Testing Guide

Guide for load testing the Job Scheduler system to validate performance and scalability.

## Objectives

Load testing should validate:
1. **API throughput**: Requests per second the API can handle
2. **Job processing capacity**: Jobs processed per hour
3. **Concurrent DAG handling**: Multiple DAGs processed simultaneously
4. **Consumer scalability**: How many consumers are needed for load
5. **Database performance**: Query performance under load
6. **Resource usage**: CPU, memory, I/O under load

## Tools

### Recommended Tools

- **k6**: Modern load testing tool for APIs
- **Apache Bench (ab)**: Simple HTTP benchmarking
- **wrk**: High-performance HTTP benchmarking
- **JMeter**: Full-featured load testing (GUI available)
- **Postman/Newman**: API testing and collection running

### Example: k6 Script

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },   // Stay at 10 users
    { duration: '30s', target: 50 },  // Ramp up to 50 users
    { duration: '2m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests < 1s
    http_req_failed: ['rate<0.01'],    // < 1% errors
  },
};

const API_URL = __ENV.API_URL || 'http://localhost:8000';
const API_KEY = __ENV.API_KEY || 'test-api-key';

export default function () {
  // Test 1: Health check
  let res = http.get(`${API_URL}/health`, {
    headers: { 'X-API-Key': API_KEY },
  });
  check(res, { 'health status is 200': (r) => r.status === 200 });

  // Test 2: Create DAG
  const uuid = `test-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  const payload = JSON.stringify([
    {
      uuid: `${uuid}-a`,
      job_title: 'Load Test Job A',
      job_type: 'test',
      data: { test: true },
      edges: [`${uuid}-b`],
    },
    {
      uuid: `${uuid}-b`,
      job_title: 'Load Test Job B',
      job_type: 'test',
      data: { test: true },
      edges: [],
    },
  ]);

  res = http.post(
    `${API_URL}/jobs/bulk-insert/complete-graph`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': API_KEY,
      },
    }
  );
  check(res, {
    'create DAG status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
```

**Run k6 test**:
```bash
k6 run load-test.js \
  --env API_URL=http://localhost:8000 \
  --env API_KEY=your-api-key
```

## Test Scenarios

### Scenario 1: API Endpoint Capacity

**Goal**: Determine maximum requests per second

**Method**:
```bash
# Using Apache Bench
ab -n 10000 -c 100 \
  -H "X-API-Key: your-api-key" \
  http://localhost:8000/health

# Using wrk
wrk -t4 -c100 -d30s \
  -H "X-API-Key: your-api-key" \
  http://localhost:8000/health
```

**Metrics to capture**:
- Requests per second
- Response time (mean, p95, p99)
- Error rate

### Scenario 2: DAG Creation Rate

**Goal**: Determine how many DAGs can be created per second

**Method**: Create DAGs continuously and measure:
- DAG creation rate
- Job insertion time
- Database write performance

**Expected**: 10-50 DAGs/second (depending on DAG size)

### Scenario 3: Concurrent Job Processing

**Goal**: Validate consumer scalability

**Method**:
1. Create large number of jobs (1000+)
2. Monitor consumer processing rate
3. Measure consumer lag
4. Scale consumers and measure improvement

**Metrics**:
- Jobs processed per second
- Consumer lag
- Processing time per job

### Scenario 4: Large DAG Processing

**Goal**: Process DAG with many jobs (100+ jobs)

**Method**:
1. Create DAG with 100+ jobs
2. Monitor execution time
3. Verify all dependencies resolve correctly
4. Measure database query performance

**Metrics**:
- Total DAG execution time
- Time per job
- Dependency resolution time

### Scenario 5: Sustained Load

**Goal**: Validate system stability under sustained load

**Method**: Run load test for extended period (1+ hours)

**Metrics**:
- Memory usage over time
- CPU usage over time
- Error rate over time
- Performance degradation

## Performance Benchmarks

### Target Metrics

Based on typical production requirements:

| Metric | Target | Notes |
|--------|--------|-------|
| API requests/second | 1000+ | With rate limiting enabled |
| DAG creation rate | 50+ DAGs/sec | Depends on DAG size |
| Job processing rate | 100+ jobs/sec | Per consumer |
| API latency (p95) | < 500ms | For health check |
| API latency (p95) | < 2s | For DAG creation |
| Consumer lag | < 1000 | Under normal load |
| Database connections | < 80% of max | Currently max 40 |

### Baseline Measurements

Measure baseline performance before scaling:

```bash
# 1. API Server
# Single instance, baseline load
k6 run api-load-test.js --vus 10 --duration 2m

# 2. Consumers
# Single consumer, measure job processing rate
# Create 1000 jobs, measure time to process

# 3. Database
# Monitor query performance
# Check slow queries: SELECT * FROM pg_stat_statements ORDER BY total_time DESC;
```

## Scaling Tests

### Test 1: API Server Horizontal Scaling

**Procedure**:
1. Start with 1 API server replica
2. Run load test, measure performance
3. Scale to 2 replicas, measure improvement
4. Scale to 3 replicas, measure improvement
5. Identify optimal replica count

**Expected**: Linear or near-linear improvement up to bottleneck (database/Kafka)

### Test 2: Consumer Horizontal Scaling

**Procedure**:
1. Create 10,000 jobs
2. Start with 1 consumer, measure processing time
3. Scale to 2 consumers, measure improvement
4. Continue scaling (3, 5, 10 consumers)
5. Identify point of diminishing returns

**Expected**: 
- Linear improvement up to partition count
- Diminishing returns after partition count (Kafka limits)

**Formula**: Optimal consumers = Kafka topic partition count

### Test 3: Database Connection Pool Tuning

**Procedure**:
1. Baseline: Default settings (MaxOpenConns=40)
2. Increase to 60, measure performance
3. Increase to 80, measure performance
4. Monitor database connection usage
5. Find optimal pool size

## Load Testing Checklist

### Pre-Test

- [ ] Test environment matches production configuration
- [ ] Database schema initialized
- [ ] Kafka topic created with appropriate partitions
- [ ] All services running and healthy
- [ ] Monitoring/observability enabled
- [ ] Baseline metrics captured

### During Test

- [ ] Monitor API request rate
- [ ] Monitor API response times
- [ ] Monitor consumer lag
- [ ] Monitor database connections
- [ ] Monitor CPU/Memory usage
- [ ] Watch for errors in logs

### Post-Test

- [ ] Analyze results
- [ ] Identify bottlenecks
- [ ] Document findings
- [ ] Recommend optimizations
- [ ] Update performance targets

## Interpreting Results

### Good Performance

- ✅ API latency (p95) < 1s
- ✅ Error rate < 0.1%
- ✅ Consumer lag < 100
- ✅ Resource usage < 70%
- ✅ No memory leaks over time

### Needs Optimization

- ⚠️ API latency (p95) > 2s
- ⚠️ Error rate > 1%
- ⚠️ Consumer lag > 1000
- ⚠️ Resource usage > 80%
- ⚠️ Performance degradation over time

### Action Items

Based on results:

1. **High API latency**: Scale API servers, optimize queries, add caching
2. **High consumer lag**: Scale consumers, increase Kafka partitions
3. **High error rate**: Investigate root cause, fix bugs, increase retries
4. **Resource exhaustion**: Scale horizontally, optimize code, increase limits
5. **Database bottlenecks**: Optimize queries, add indexes, increase connection pool

## Example Load Test Results

### Test Configuration

- API Servers: 2 replicas
- Consumers: 3 replicas
- Kafka Partitions: 6
- Test Duration: 30 minutes
- Load: 50 concurrent users

### Results

| Metric | Result | Status |
|--------|--------|--------|
| API requests/sec | 850 | ✅ Good |
| API latency (p95) | 450ms | ✅ Good |
| API error rate | 0.05% | ✅ Good |
| Jobs processed/sec | 120 | ✅ Good |
| Consumer lag (max) | 150 | ✅ Good |
| Database connections | 25/40 | ✅ Good |
| CPU usage (avg) | 45% | ✅ Good |
| Memory usage (avg) | 60% | ✅ Good |

### Conclusion

System performs well under load. Can handle:
- 850 requests/second
- 120 jobs/second processing
- 50 concurrent users

**Recommendations**:
- Current configuration sufficient for expected load
- Can scale consumers to 6 (match partition count) for higher throughput
- Monitor for sustained load over longer periods

## References

- k6 Documentation: https://k6.io/docs/
- Apache Bench: https://httpd.apache.org/docs/2.4/programs/ab.html
- wrk: https://github.com/wg/wrk
