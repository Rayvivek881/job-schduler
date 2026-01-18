# Distributed DAG Job Scheduler

A high-performance, horizontally scalable job scheduler built in Go that supports **Directed Acyclic Graph (DAG)** based job dependencies with distributed parallel execution using Kafka.

---

## 🎯 Key Features

- **DAG-based Job Dependencies** - Jobs can depend on other jobs, forming a directed acyclic graph
- **Distributed Parallel Execution** - Multiple workers process jobs concurrently via Kafka consumer groups
- **Degree-based Scheduling** - Elegant algorithm using in-degree to determine job readiness
- **Automatic Dependency Resolution** - Dependent jobs auto-trigger when predecessors complete
- **Retry Mechanism** - Configurable retry count and intervals with exponential backoff support
- **Job-level Error Tracking** - Runtime errors persisted in PostgreSQL JSONB field
- **Rate Limiting** - Token bucket algorithm for API protection
- **Clean Architecture** - Layered design with clear separation of concerns

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT REQUEST                                  │
│                         (POST /jobs/bulk-insert)                            │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              API SERVER                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────────────┐  │
│  │Rate Limiter │─▶│ Auth Guard  │─▶│  DAG Validator (Topological Sort)   │  │
│  │(Token Bucket)│  │ (API Key)   │  │  - Cycle Detection                  │  │
│  └─────────────┘  └─────────────┘  │  - In-Degree Calculation            │  │
│                                     └─────────────────────────────────────┘  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            POSTGRESQL                                        │
│  ┌─────────────────────────────┐  ┌─────────────────────────────────────┐   │
│  │         jobs Table          │  │          edges Table                │   │
│  │  - uuid (PK)                │  │  - source (FK → jobs.uuid)          │   │
│  │  - degree (in-degree)       │  │  - target (FK → jobs.uuid)          │   │
│  │  - status                   │  │                                     │   │
│  │  - data (JSONB)             │  │  Represents: source ──▶ target      │   │
│  │  - job_response (JSONB)     │  │  (source must complete before       │   │
│  │  - retry_count              │  │   target can run)                   │   │
│  │  - run_after                │  │                                     │   │
│  └─────────────────────────────┘  └─────────────────────────────────────┘   │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                             SCHEDULER                                        │
│                                                                              │
│    SELECT * FROM jobs WHERE degree = 0 AND status = 'open'                  │
│                           AND run_after <= NOW()                            │
│                                                                              │
│    Jobs with degree 0 = No dependencies = Ready to execute                  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              KAFKA                                           │
│                         (jobs-topic)                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Partition 0  │  Partition 1  │  Partition 2  │  Partition N  │    │    │
│  │    [Job A]    │    [Job B]    │    [Job C]    │    [Job X]    │    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
            ┌─────────────────────┼─────────────────────┐
            │                     │                     │
            ▼                     ▼                     ▼
┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐
│    Consumer 1     │ │    Consumer 2     │ │    Consumer N     │
│    (Worker)       │ │    (Worker)       │ │    (Worker)       │
│                   │ │                   │ │                   │
│  - Fetch job      │ │  - Fetch job      │ │  - Fetch job      │
│  - Process        │ │  - Process        │ │  - Process        │
│  - Update status  │ │  - Update status  │ │  - Update status  │
│  - Decrement      │ │  - Decrement      │ │  - Decrement      │
│    dependents'    │ │    dependents'    │ │    dependents'    │
│    degree         │ │    degree         │ │    degree         │
└───────────────────┘ └───────────────────┘ └───────────────────┘
            │                     │                     │
            └─────────────────────┴─────────────────────┘
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │  UPDATE jobs           │
                    │  SET degree = degree-1 │
                    │  WHERE uuid IN (       │
                    │    SELECT target       │
                    │    FROM edges          │
                    │    WHERE source = ?    │
                    │  )                     │
                    └─────────────────────────┘
                                  │
                                  ▼
                    ┌─────────────────────────┐
                    │  Next scheduler tick:   │
                    │  Newly degree-0 jobs    │
                    │  become eligible        │
                    └─────────────────────────┘
```

---

## 🔄 The Degree-Based DAG Execution Algorithm

### How It Works

This scheduler implements **Kahn's Algorithm** adapted for distributed job execution:

#### Step 1: Job Insertion with DAG Validation

```
Input Graph:
    A ──→ B ──→ D
          │
          └──→ C

Calculated In-Degrees:
    A: 0 (no dependencies)
    B: 1 (depends on A)
    C: 1 (depends on B)
    D: 1 (depends on B)
```

When jobs are inserted via API:
1. Build adjacency list from edges
2. Calculate in-degree for each node using topological sort
3. Detect cycles (if found, reject the request)
4. Store jobs with their calculated `degree` value

#### Step 2: Scheduler Picks Ready Jobs

```sql
SELECT * FROM jobs 
WHERE degree = 0          -- No pending dependencies
  AND status = 'open'     -- Not already processing
  AND run_after <= NOW()  -- Scheduled time has passed
```

Only jobs with `degree = 0` are pushed to Kafka.

#### Step 3: Parallel Processing

Multiple consumers (workers) pull jobs from Kafka:
- Each consumer belongs to the same **consumer group**
- Kafka ensures each job is processed by exactly one consumer
- Workers process jobs in parallel across the cluster

#### Step 4: Dependency Cascade

When a job completes:

```sql
UPDATE jobs 
SET degree = degree - 1 
WHERE uuid IN (
    SELECT target FROM edges WHERE source = 'completed-job-uuid'
)
```

This decrements the degree of all dependent jobs. When their degree reaches 0, they become eligible for the next scheduler tick.

#### Visual Example

```
Time T0: Initial State
┌─────────────────────────────────────────────────┐
│  Job A (degree: 0) ✓ Ready                      │
│  Job B (degree: 1) ⏳ Waiting for A             │
│  Job C (degree: 1) ⏳ Waiting for B             │
│  Job D (degree: 1) ⏳ Waiting for B             │
└─────────────────────────────────────────────────┘

Time T1: Job A Completes
┌─────────────────────────────────────────────────┐
│  Job A (degree: 0) ✅ Completed                 │
│  Job B (degree: 0) ✓ Ready (decremented)        │
│  Job C (degree: 1) ⏳ Waiting for B             │
│  Job D (degree: 1) ⏳ Waiting for B             │
└─────────────────────────────────────────────────┘

Time T2: Job B Completes
┌─────────────────────────────────────────────────┐
│  Job A (degree: 0) ✅ Completed                 │
│  Job B (degree: 0) ✅ Completed                 │
│  Job C (degree: 0) ✓ Ready (decremented)        │
│  Job D (degree: 0) ✓ Ready (decremented)        │
└─────────────────────────────────────────────────┘

Time T3: Jobs C & D Process in Parallel
┌─────────────────────────────────────────────────┐
│  Consumer 1 processes Job C                     │
│  Consumer 2 processes Job D    (parallel!)      │
└─────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
job-scheduler/
├── cmd/                        # CLI Commands (Cobra)
│   ├── root.go                 # Root command
│   ├── server.go               # API server command
│   ├── scheduler.go            # Job scheduler command
│   └── consumer.go             # Kafka consumer command
│
├── clients/                    # External Service Clients
│   ├── kafka.go                # Sarama Kafka wrapper
│   └── pgsql.go                # PostgreSQL client (Bun ORM)
│
├── conf/                       # Configuration
│   └── viper.go                # Viper config loader
│
├── connections/                # Global Connection Instances
│   ├── database.go             # PostgreSQL connection
│   └── kafka.go                # Kafka service instance
│
├── constants/                  # Constants & Errors
│   ├── errors_message.go       # Centralized error definitions
│   └── jobs.go                 # Job status constants
│
├── jobs/                       # Job Processing Logic
│   ├── schduler.go             # Scheduler (pushes to Kafka)
│   └── consumers/
│       └── base.go             # Kafka consumer implementation
│
├── middleware/                 # HTTP Middleware
│   ├── authMiddleware.go       # API key authentication
│   └── rateMiddleware.go       # Token bucket rate limiter
│
├── models/                     # Data Models & Repositories
│   ├── job_node.go             # Job model
│   ├── job_node.repo.go        # Job repository
│   ├── edge.go                 # Edge model (dependencies)
│   └── edge.repo.go            # Edge repository
│
├── server/                     # HTTP Server
│   ├── routes.go               # Route definitions
│   ├── controller/             # Request handlers
│   ├── helper/                 # Request validation & DAG logic
│   └── services/               # Business logic layer
│
├── utilities/                  # Helper Functions
│   ├── comman.go               # Common utilities
│   └── structure.go            # Pagination helpers
│
└── main.go                     # Application entry point
```

---

## 🚀 Running the Application

### Prerequisites

- Go 1.24+
- PostgreSQL 14+
- Apache Kafka 3.x

### Configuration

Create a `.env` file:

```env
# Application
APP_ENV=development
API_KEY=your-secret-api-key
MAX_REQUESTS_PER_MINUTE=1000

# PostgreSQL
PG_DB_HOST=localhost
PG_DB_PORT=5432
PG_DB_DATABASE=job_scheduler
PG_DB_USERNAME=postgres
PG_DB_PASSWORD=postgres
PG_DB_DEBUG=true

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_VERSION=3.6.0
JOBS_TOPIC=jobs-topic
CONSUMER_GROUP=jobs-consumer-group

# Job Configuration
TICKER_INTERVAL=1
MAX_REQUESTS_PER_MINUTE=1000
```

### Commands

```bash
# Start API Server
go run . api-server

# Start Scheduler (pushes ready jobs to Kafka)
go run . scheduler first-time

# Start Retry Scheduler (retries failed jobs)
go run . scheduler retry

# Start Consumer (processes jobs from Kafka)
go run . consumer
```

### For Horizontal Scaling

```bash
# Run multiple consumers (each joins the same consumer group)
go run . consumer &
go run . consumer &
go run . consumer &

# Kafka automatically distributes partitions among consumers
```

---

## 📡 API Endpoints

### Insert Jobs with Dependencies

```bash
POST /jobs/bulk-insert/complete-graph
Content-Type: application/json
X-API-Key: your-api-key

[
  {
    "uuid": "job-a",
    "job_title": "Extract Data",
    "job_type": "etl",
    "data": {"source": "s3://bucket/data"},
    "retry_count": 3,
    "retry_interval": 5,
    "edges": ["job-b"]  // job-a must complete before job-b
  },
  {
    "uuid": "job-b",
    "job_title": "Transform Data",
    "job_type": "etl",
    "data": {"transform": "aggregate"},
    "retry_count": 3,
    "retry_interval": 5,
    "edges": ["job-c", "job-d"]
  },
  {
    "uuid": "job-c",
    "job_title": "Load to Warehouse",
    "job_type": "etl",
    "data": {"destination": "warehouse"},
    "edges": []
  },
  {
    "uuid": "job-d",
    "job_title": "Generate Report",
    "job_type": "report",
    "data": {"format": "pdf"},
    "edges": []
  }
]
```

### Get Jobs with Filters

```bash
GET /jobs/?status=open&status=failed&page=1&limit=10
X-API-Key: your-api-key
```

### Retry a Failed Job

```bash
PUT /jobs/{uuid}/retry
Content-Type: application/json
X-API-Key: your-api-key

{
  "data": {"updated": "payload"},
  "retry_count": 5
}
```

### Health Check

```bash
GET /health
```

---

## 🔧 Technical Highlights

### 1. Token Bucket Rate Limiting

```go
// Refills tokens every second
// Configurable via MAX_REQUESTS_PER_MINUTE
func RateLimiter() gin.HandlerFunc {
    tokens = maxLimit
    fillingRate = maxLimit / 60
    
    // Background refiller
    go func() {
        for range ticker.C {
            tokens = min(maxLimit, tokens + fillingRate)
        }
    }()
}
```

### 2. Centralized Error Handling

```go
// All errors defined in one place
var (
    ErrJobNotFound      = errors.New("job operation failed: job not found")
    ErrJobStatusInvalid = errors.New("job operation failed: invalid status")
    ErrKafkaWrite       = errors.New("kafka write failed")
)

// Consistent error wrapping
return constants.ErrorWrap(constants.ErrJobFetch, err)
```

### 3. Async Kafka Producer

```go
// Non-blocking message publishing
func (w *KafkaWriter) BulkWrite(values any) error {
    for each value {
        w.producer.Input() <- &sarama.ProducerMessage{
            Topic: w.topic,
            Key:   uuid.NewString(),
            Value: jsonMarshal(value),
        }
    }
    return nil  // Returns immediately
}
```

### 4. Job Error Persistence

```go
// Errors stored in JSONB for querying
type JobResponseStruct struct {
    Message       string   `json:"message"`
    RuntimeErrors []string `json:"runtime_errors"`
}

// On failure
job.AddToJobResponse("runtime_errors", err.Error())
```

---

## 📊 Job Lifecycle

```
┌──────────┐     ┌──────────┐     ┌────────────┐     ┌───────────┐
│   open   │────▶│ in_queue │────▶│ processing │────▶│ completed │
└──────────┘     └──────────┘     └────────────┘     └───────────┘
     │                                   │
     │                                   ▼
     │                            ┌──────────┐
     └───────────────────────────▶│  failed  │
           (on error)             └──────────┘
                                       │
                                       ▼
                              (retry scheduler picks up
                               when run_after <= now)
```

---

## 🎯 Use Cases

1. **ETL Pipelines** - Extract → Transform → Load with dependencies
2. **CI/CD Pipelines** - Build → Test → Deploy stages
3. **Data Processing** - Multi-step data transformations
4. **Report Generation** - Aggregate → Calculate → Generate
5. **Workflow Automation** - Any multi-step process with dependencies

---

## 📈 Scalability

| Component | Scaling Strategy |
|-----------|------------------|
| API Server | Horizontal (load balancer) |
| Scheduler | Single instance (leader election possible) |
| Consumers | Horizontal (Kafka consumer groups) |
| PostgreSQL | Read replicas, connection pooling |
| Kafka | Add partitions for parallelism |

---

## 🛠️ Tech Stack

| Technology | Purpose |
|------------|---------|
| **Go 1.24** | Primary language |
| **Gin** | HTTP framework |
| **Cobra** | CLI framework |
| **Bun** | PostgreSQL ORM |
| **Sarama** | Kafka client (IBM) |
| **Viper** | Configuration |
| **Zerolog** | Structured logging |

---

## 📝 License

MIT License

---

## 👤 Author

Built as a demonstration of distributed systems design with DAG-based job scheduling.

