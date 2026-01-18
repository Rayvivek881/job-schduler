# Job Scheduler API Documentation

Complete API reference for the Distributed DAG Job Scheduler service.

## Base URL

- **Development**: `http://localhost:8000`
- **Production**: Configure via environment variables

## Authentication

All endpoints except `/health` require API key authentication via `X-API-Key` header.

```http
X-API-Key: your-secret-api-key
```

## Rate Limiting

API is rate-limited using token bucket algorithm:
- **Default**: 1000 requests per minute
- **Configurable**: Via `MAX_REQUESTS_PER_MINUTE` environment variable
- **Response**: `429 Too Many Requests` when limit exceeded

---

## Endpoints

### System

#### Health Check

```http
GET /health
```

Check API availability. No authentication required.

**Response** (200 OK):
```json
{
  "status": "ok"
}
```

---

### Jobs

#### Bulk Insert Complete Graph

```http
POST /jobs/bulk-insert/complete-graph
Content-Type: application/json
X-API-Key: your-api-key
```

Insert jobs with DAG dependencies. Creates jobs and edges in a single transaction with DAG validation (cycle detection and in-degree calculation).

**Request Body**: Array of job nodes
```json
[
  {
    "uuid": "job-a",
    "job_title": "Extract Data",
    "job_type": "etl",
    "data": {
      "source": "s3://bucket/data",
      "format": "csv"
    },
    "retry_count": 3,
    "retry_interval": 5,
    "edges": ["job-b"]
  },
  {
    "uuid": "job-b",
    "job_title": "Transform Data",
    "job_type": "etl",
    "data": {
      "transform": "aggregate"
    },
    "retry_count": 3,
    "retry_interval": 5,
    "edges": []
  }
]
```

**Request Fields**:
- `uuid` (string, required): Unique job identifier
- `job_title` (string, required): Human-readable job name
- `job_type` (string, required): Type of job (e.g., "etl", "report")
- `data` (object, optional): Job payload as JSONB
- `retry_count` (int, optional): Number of retries (default: 0)
- `retry_interval` (int, optional): Minutes between retries (default: 30)
- `edges` (array, optional): Array of target UUIDs (dependencies)

**Response** (200 OK):
```json
{
  "message": "Jobs inserted successfully"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid DAG (cycles detected or invalid structure)
- `400 Bad Request`: Duplicate job UUIDs
- `401 Unauthorized`: Invalid or missing API key
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Database or processing error

**Constraints**:
- Maximum 5000 nodes per request (configurable via `MaxNodesPerRequest`)
- DAG must be acyclic (no circular dependencies)
- All edge targets must exist in the same request

---

#### Get Jobs

```http
GET /jobs/?status=open&status=processing&page=1&limit=50
X-API-Key: your-api-key
```

Get jobs with optional filters. Supports filtering by UUID, status, degree, run_after, and pagination.

**Query Parameters**:
- `uuid` (array, optional): Filter by job UUIDs (multiple values supported)
- `status` (array, optional): Filter by status values (multiple values supported)
  - Valid values: `open`, `in_queue`, `processing`, `completed`, `failed`
- `degree` (int, optional): Filter by in-degree count
- `run_after` (timestamp, optional): Filter jobs scheduled after timestamp (ISO 8601 format)
- `page` (int, optional): Page number for pagination (default: 1, 1-indexed)
- `limit` (int, optional): Number of results per page (default: 100, max: 100)

**Example**:
```http
GET /jobs/?status=processing&status=completed&page=1&limit=20
GET /jobs/?uuid=job-a&uuid=job-b
GET /jobs/?degree=0&status=open
GET /jobs/?status=failed&run_after=2024-01-15T10:00:00Z
```

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": 1,
      "uuid": "job-a",
      "job_title": "Extract Data",
      "job_type": "etl",
      "status": "processing",
      "degree": 0,
      "data": {
        "source": "s3://bucket/data"
      },
      "job_response": {
        "message": "Job started"
      },
      "retry_count": 3,
      "retry_interval": 5,
      "run_after": "2024-01-15T10:00:00Z",
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T10:05:00Z"
    }
  ]
}
```

---

#### Get Job by UUID

```http
GET /jobs/:uuid
X-API-Key: your-api-key
```

Get a single job by UUID.

**Path Parameters**:
- `uuid` (string, required): Job UUID

**Response** (200 OK):
```json
{
  "data": {
    "id": 1,
    "uuid": "job-a",
    "job_title": "Extract Data",
    "job_type": "etl",
    "status": "completed",
    "degree": 0,
    "data": {...},
    "job_response": {...},
    "retry_count": 3,
    "retry_interval": 5,
    "run_after": "2024-01-15T10:00:00Z",
    "created_at": "2024-01-15T09:00:00Z",
    "updated_at": "2024-01-15T10:15:00Z"
  }
}
```

**Error Responses**:
- `404 Not Found`: Job not found
- `401 Unauthorized`: Invalid or missing API key

---

#### Update and Retry Job

```http
PUT /jobs/:uuid/retry
Content-Type: application/json
X-API-Key: your-api-key
```

Update a job and retrigger it. Updates job data and/or retry count, then sets status to 'open' and resets run_after to current time.

**Path Parameters**:
- `uuid` (string, required): Job UUID

**Request Body**:
```json
{
  "data": {
    "updated": "payload",
    "source": "s3://bucket/new-data"
  },
  "retry_count": 5
}
```

**Request Fields**:
- `data` (object, optional): Updated job payload
- `retry_count` (int, optional): Updated retry count

**Response** (200 OK):
```json
{
  "message": "Job updated and retriggered successfully"
}
```

**Error Responses**:
- `404 Not Found`: Job not found
- `400 Bad Request`: Invalid request body
- `401 Unauthorized`: Invalid or missing API key

---

### DAG

#### Get DAG Status

```http
GET /jobs/dag/status?uuids=job-a&uuids=job-b&uuids=job-c
X-API-Key: your-api-key
```

Get aggregated DAG status for multiple jobs. Returns overall DAG status, status breakdown, individual job statuses, and progress metrics.

**Query Parameters**:
- `uuids` (array, required): Job UUIDs to include in DAG status calculation (multiple values required)

**Response** (200 OK):
```json
{
  "dag_status": "processing",
  "total_jobs": 4,
  "status_breakdown": {
    "completed": 2,
    "processing": 1,
    "open": 1
  },
  "jobs": {
    "job-a": {
      "uuid": "job-a",
      "job_title": "Extract Data",
      "status": "completed",
      "degree": 0,
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T10:15:00Z"
    },
    "job-b": {
      "uuid": "job-b",
      "job_title": "Transform Data",
      "status": "processing",
      "degree": 0,
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T10:20:00Z"
    }
  },
  "progress": {
    "completed": 2,
    "failed": 0,
    "processing": 1,
    "pending": 1,
    "percentage": 50.0
  }
}
```

**DAG Status Values**:
- `completed`: All jobs completed
- `failed`: One or more jobs failed
- `processing`: One or more jobs are processing or in queue
- `pending`: All jobs are open (not started)
- `partial`: Mixed states

**Error Responses**:
- `400 Bad Request`: No UUIDs provided
- `401 Unauthorized`: Invalid or missing API key

---

#### Get DAG Progress

```http
GET /jobs/dag/progress?uuids=job-a&uuids=job-b&uuids=job-c
X-API-Key: your-api-key
```

Get DAG progress summary with completion percentage. Lightweight version of DAG status focused on progress metrics.

**Query Parameters**:
- `uuids` (array, required): Job UUIDs to include in progress calculation (multiple values required)

**Response** (200 OK):
```json
{
  "progress": {
    "completed": 2,
    "failed": 0,
    "processing": 1,
    "pending": 1,
    "percentage": 50.0
  },
  "status_breakdown": {
    "completed": 2,
    "processing": 1,
    "open": 1
  },
  "dag_status": "processing"
}
```

**Error Responses**:
- `400 Bad Request`: No UUIDs provided
- `401 Unauthorized`: Invalid or missing API key

---

## Job Status Values

| Status | Description |
|--------|-------------|
| `open` | Job created, waiting to be picked up by scheduler |
| `in_queue` | Job queued in Kafka, waiting for consumer |
| `processing` | Job currently being executed by consumer |
| `completed` | Job finished successfully |
| `failed` | Job execution failed (can be retried) |

## Job Lifecycle

```
open → in_queue → processing → completed
                        ↓
                     failed → (retry) → open
```

---

## Error Codes

| Code | Description |
|------|-------------|
| `400 Bad Request` | Invalid request format or invalid DAG |
| `401 Unauthorized` | Invalid or missing API key |
| `404 Not Found` | Job not found |
| `429 Too Many Requests` | Rate limit exceeded |
| `500 Internal Server Error` | Server error (database, Kafka, etc.) |

---

## Examples

### Creating an ETL Pipeline DAG

```bash
curl -X POST "http://localhost:8000/jobs/bulk-insert/complete-graph" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '[
    {
      "uuid": "extract-2024-01-15",
      "job_title": "Extract Data",
      "job_type": "etl",
      "data": {"source_bucket": "my-bucket", "source_key": "data/2024/01/15/input.csv"},
      "retry_count": 3,
      "edges": ["transform-2024-01-15"]
    },
    {
      "uuid": "transform-2024-01-15",
      "job_title": "Transform Data",
      "job_type": "etl",
      "data": {"transformations": ["clean", "normalize"]},
      "retry_count": 3,
      "edges": ["load-2024-01-15"]
    },
    {
      "uuid": "load-2024-01-15",
      "job_title": "Load to Warehouse",
      "job_type": "etl",
      "data": {"destination": "warehouse", "table": "daily_data"},
      "edges": []
    }
  ]'
```

### Checking DAG Status

```bash
curl -X GET "http://localhost:8000/jobs/dag/status?uuids=extract-2024-01-15&uuids=transform-2024-01-15&uuids=load-2024-01-15" \
  -H "X-API-Key: your-api-key"
```

### Querying Failed Jobs

```bash
curl -X GET "http://localhost:8000/jobs/?status=failed&page=1&limit=50" \
  -H "X-API-Key: your-api-key"
```

---

## See Also

- [DAG Guide](./DAG_GUIDE.md) - How to create and manage DAGs
- [Configuration Guide](./CONFIGURATION.md) - Environment variables and configuration
- [README.md](../README.md) - Project overview and architecture
