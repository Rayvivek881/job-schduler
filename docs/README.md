# Job Scheduler Documentation

Complete documentation for the Distributed DAG Job Scheduler API service.

## Documentation Index

### Getting Started

- **[API Documentation](./API.md)** - Complete API reference with endpoints, request/response formats, and examples
- **[DAG Guide](./DAG_GUIDE.md)** - How to create and manage DAGs with dependencies
- **[Configuration Guide](./CONFIGURATION.md)** - Environment variables and configuration options

### Additional Resources

- **[Main README](../README.md)** - Project overview, architecture, and getting started
- **[Postman Collection](../postman/Job Scheduler API.postman_collection.json)** - Postman collection for API testing

## Quick Links

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check (no auth required) |
| `/jobs/bulk-insert/complete-graph` | POST | Insert jobs with DAG dependencies |
| `/jobs/` | GET | Get jobs with filters |
| `/jobs/:uuid` | GET | Get single job by UUID |
| `/jobs/:uuid/retry` | PUT | Update and retrigger job |
| `/jobs/dag/status` | GET | Get aggregated DAG status |
| `/jobs/dag/progress` | GET | Get DAG progress with percentage |

### Key Concepts

- **DAG (Directed Acyclic Graph)**: Collection of jobs with dependencies
- **Degree (In-degree)**: Number of dependencies a job has
- **Job Status**: `open`, `in_queue`, `processing`, `completed`, `failed`
- **Processor**: Job execution handler (default or custom)

## Quick Start

### 1. Configuration

Create `.env` file:

```env
API_KEY=your-secret-api-key
PG_DB_HOST=localhost
PG_DB_DATABASE=job_scheduler
PG_DB_USERNAME=postgres
PG_DB_PASSWORD=postgres
KAFKA_BROKERS=localhost:9092
KAFKA_VERSION=3.6.0
JOBS_TOPIC=jobs-topic
CONSUMER_GROUP=jobs-consumer-group
```

See [Configuration Guide](./CONFIGURATION.md) for complete reference.

### 2. Create a DAG

```bash
curl -X POST "http://localhost:8000/jobs/bulk-insert/complete-graph" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '[
    {
      "uuid": "job-a",
      "job_title": "Extract Data",
      "job_type": "etl",
      "data": {"source": "s3://bucket/data"},
      "edges": ["job-b"]
    },
    {
      "uuid": "job-b",
      "job_title": "Transform Data",
      "job_type": "etl",
      "data": {"transform": "aggregate"},
      "edges": []
    }
  ]'
```

See [DAG Guide](./DAG_GUIDE.md) for detailed examples.

### 3. Check Status

```bash
# Get job status
curl -X GET "http://localhost:8000/jobs/job-a" \
  -H "X-API-Key: your-api-key"

# Get DAG status
curl -X GET "http://localhost:8000/jobs/dag/status?uuids=job-a&uuids=job-b" \
  -H "X-API-Key: your-api-key"
```

See [API Documentation](./API.md) for complete endpoint reference.

## Features

- **DAG-based Job Dependencies**: Create complex workflows with automatic dependency resolution
- **Distributed Parallel Execution**: Multiple workers process jobs concurrently via Kafka
- **Job Status Tracking**: Query and monitor job execution status
- **DAG Status Aggregation**: Monitor overall DAG progress and completion
- **Retry Mechanism**: Automatic retry with configurable intervals
- **Rate Limiting**: Token bucket algorithm for API protection
- **Horizontal Scaling**: Add consumers for parallel processing

## Examples

### ETL Pipeline

```
Extract → Transform → Load
```

### Parallel Processing

```
Split → [Job1, Job2, Job3] → Merge
```

### Sequential Steps

```
Step1 → Step2 → Step3 → Step4
```

See [DAG Guide](./DAG_GUIDE.md) for detailed examples and patterns.

## Support

For issues, questions, or contributions:

1. Check [API Documentation](./API.md) for endpoint details
2. Review [DAG Guide](./DAG_GUIDE.md) for DAG creation patterns
3. Consult [Configuration Guide](./CONFIGURATION.md) for environment setup
4. See [Main README](../README.md) for architecture and troubleshooting

---

**Last Updated**: 2024-01-15  
**Version**: 1.0.0
