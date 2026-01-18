# Complete Guide: How to Create DAGs

This comprehensive guide explains all methods for creating DAGs (Directed Acyclic Graphs) in the Job Scheduler system.

---

## 📚 Table of Contents

1. [Understanding DAGs](#understanding-dags)
2. [Three Ways to Create DAGs](#three-ways-to-create-dags)
3. [Method 1: Manual JSON (Direct API)](#method-1-manual-json-direct-api)
4. [Method 2: TypeScript Client Library](#method-2-typescript-client-library)
5. [Method 3: Python Client Library](#method-3-python-client-library)
6. [DAG Structure Deep Dive](#dag-structure-deep-dive)
7. [Dependency Patterns](#dependency-patterns)
8. [DAG Validation](#dag-validation)
9. [Complete Examples](#complete-examples)
10. [Task Breakdown](#task-breakdown)

---

## Understanding DAGs

### What is a DAG?

A **Directed Acyclic Graph (DAG)** is a collection of jobs with dependencies where:
- **Directed**: Dependencies have direction (A → B means A must complete before B)
- **Acyclic**: No circular dependencies (no cycles like A → B → C → A)
- **Graph**: Jobs are nodes, dependencies are edges

### Key Concepts

| Concept | Description | Example |
|---------|-------------|---------|
| **Node (Job)** | A task to be executed | "Extract Data", "Transform Data" |
| **Edge (Dependency)** | Relationship between jobs | `edges: ["job-b"]` means this job depends on `job-b` |
| **Degree (In-degree)** | Number of dependencies a job has | `degree: 0` = ready to execute |
| **Topological Sort** | Algorithm that determines execution order | Kahn's Algorithm |

### Visual Example

```
┌─────────────────────────────────────┐
│          DAG Example                │
├─────────────────────────────────────┤
│                                     │
│  Job A (degree: 0) ✓ Ready         │
│    │                                │
│    ├──────────┬─────────────────┐   │
│    │          │                 │   │
│    ▼          ▼                 ▼   │
│  Job B     Job C             Job D  │
│  (degree: 1) (degree: 1)  (degree: 1) ⏳ Waiting
│    │          │                 │   │
│    └──────┬───┘                 │   │
│           │                     │   │
│           ▼                     ▼   │
│         Job E                 Job F │
│      (degree: 2)          (degree: 2) ⏳ Waiting
│                                     │
└─────────────────────────────────────┘
```

**Execution Flow:**
1. Job A executes first (degree: 0)
2. Job A completes → Job B, C, D degrees decrement to 0
3. Job B, C, D execute in parallel (all degree: 0)
4. Job B, C complete → Job E, F degrees decrement to 0
5. Job E, F execute in parallel (both degree: 0)

---

## Three Ways to Create DAGs

### Overview

You can create DAGs using three methods:

1. **Manual JSON** - Direct API call with manually constructed JSON
2. **TypeScript Library** - Type-safe builders with IntelliSense
3. **Python Library** - Python builders with validation

### Comparison

| Method | Pros | Cons | Best For |
|--------|------|------|----------|
| **Manual JSON** | Full control, no dependencies | Error-prone, verbose | Simple DAGs, testing |
| **TypeScript** | Type-safe, IntelliSense, validation | Requires TypeScript setup | TypeScript/JavaScript projects |
| **Python** | Simple, readable, validation | Requires Python environment | Python projects, scripts |

---

## Method 1: Manual JSON (Direct API)

### API Endpoint

```
POST /jobs/bulk-insert/complete-graph
```

### Headers

```json
{
  "Content-Type": "application/json",
  "X-API-Key": "your-api-key-here"
}
```

### Request Body Format

```json
[
  {
    "uuid": "job-uuid-1",
    "job_title": "Human Readable Job Name",
    "job_type": "job_type_string",
    "data": {
      "key": "value",
      "custom_vars": {}
    },
    "retry_count": 3,
    "retry_interval": 5,
    "edges": ["job-uuid-2", "job-uuid-3"]
  },
  {
    "uuid": "job-uuid-2",
    "job_title": "Another Job",
    "job_type": "job_type_string",
    "data": {},
    "edges": []
  }
]
```

### Example: Simple Sequential DAG

```bash
curl -X POST http://localhost:8000/jobs/bulk-insert/complete-graph \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '[
    {
      "uuid": "extract-job",
      "job_title": "Extract Data from S3",
      "job_type": "etl",
      "data": {
        "source": "s3://bucket/data.csv",
        "format": "csv"
      },
      "retry_count": 3,
      "retry_interval": 5,
      "edges": ["transform-job"]
    },
    {
      "uuid": "transform-job",
      "job_title": "Transform Data",
      "job_type": "etl",
      "data": {
        "transformations": ["clean", "normalize"]
      },
      "retry_count": 3,
      "retry_interval": 5,
      "edges": ["load-job"]
    },
    {
      "uuid": "load-job",
      "job_title": "Load to Warehouse",
      "job_type": "etl",
      "data": {
        "destination": "warehouse",
        "table": "daily_data"
      },
      "retry_count": 3,
      "retry_interval": 5,
      "edges": []
    }
  ]'
```

### Example: Parallel Fan-out DAG

```json
[
  {
    "uuid": "split-job",
    "job_title": "Split Data",
    "job_type": "etl",
    "data": {"operation": "split"},
    "edges": ["process-1", "process-2", "process-3"]
  },
  {
    "uuid": "process-1",
    "job_title": "Process Batch 1",
    "job_type": "etl",
    "data": {"batch": 1},
    "edges": ["merge-job"]
  },
  {
    "uuid": "process-2",
    "job_title": "Process Batch 2",
    "job_type": "etl",
    "data": {"batch": 2},
    "edges": ["merge-job"]
  },
  {
    "uuid": "process-3",
    "job_title": "Process Batch 3",
    "job_type": "etl",
    "data": {"batch": 3},
    "edges": ["merge-job"]
  },
  {
    "uuid": "merge-job",
    "job_title": "Merge Results",
    "job_type": "etl",
    "data": {"operation": "merge"},
    "edges": []
  }
]
```

**Execution**: `split-job` → `process-1, process-2, process-3` (parallel) → `merge-job`

---

## Method 2: TypeScript Client Library

### Installation

```bash
cd client-libs/typescript
npm install
npm run build
```

### Import Builders

```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";
import { createEmailExportDAG } from "./src/builders/EmailExportDAGBuilder";
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";
```

### Example 1: Contact Import DAG

```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

// Basic usage
const dag = createContactImportDAG(
  "my-uploads-bucket",
  "uploads/contacts_2024.csv"
);

// With options
const dagWithOptions = createContactImportDAG(
  "my-uploads-bucket",
  "uploads/contacts_2024.csv",
  {
    workflowId: "custom-workflow-id",
    retryCount: 5,
    retryInterval: 10,
    customVars: {
      source: "manual_upload",
      user_id: "user-123",
      campaign_id: "campaign-456"
    }
  }
);

// Submit to API
const response = await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dag)
});

const result = await response.json();
console.log(result);
```

### Example 2: Email Export DAG

```typescript
import { createEmailExportDAG } from "./src/builders/EmailExportDAGBuilder";

const dag = createEmailExportDAG(
  "my-exports-bucket",
  "exports/contacts.csv",
  {
    outputS3Key: "exports/enriched/contacts_enriched.csv",
    emailApiUrl: "https://api.emailfinder.com/v1",
    emailApiKey: "api-key-123",
    batchSize: 100,
    retryCount: 5,
    retryInterval: 10
  }
);

// Submit to API
await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dag)
});
```

### Example 3: Custom DAG Using BaseDAGBuilder

```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

// Generate workflow ID
const workflowId = BaseDAGBuilder.generateWorkflowId("my-workflow");

// Build custom job nodes
const dag = [
  BaseDAGBuilder.buildJobNode(
    BaseDAGBuilder.generateUUID("extract"),
    "Extract Data",
    "etl",
    {
      source: "s3://bucket/data.csv",
      format: "csv",
      custom_vars: {
        date: "2024-01-15",
        region: "us-east-1"
      }
    },
    {
      edges: [BaseDAGBuilder.generateUUID("transform")],
      retryCount: 3,
      retryInterval: 5
    }
  ),
  BaseDAGBuilder.buildJobNode(
    BaseDAGBuilder.generateUUID("transform"),
    "Transform Data",
    "etl",
    {
      transformations: ["clean", "normalize"]
    },
    {
      edges: [],
      retryCount: 3,
      retryInterval: 5
    }
  )
];

// Validate before submission
const validation = BaseDAGBuilder.validateDAG(dag);
if (!validation.isValid) {
  throw new Error(`DAG validation failed: ${validation.error}`);
}

// Submit to API
await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dag)
});
```

### Example 4: Complex DAG with Multiple Dependencies

```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

// Generate UUIDs
const uuidA = BaseDAGBuilder.generateUUID("job-a");
const uuidB = BaseDAGBuilder.generateUUID("job-b");
const uuidC = BaseDAGBuilder.generateUUID("job-c");
const uuidD = BaseDAGBuilder.generateUUID("job-d");

const dag = [
  // Job A (source, no dependencies)
  BaseDAGBuilder.buildJobNode(
    uuidA,
    "Extract Data",
    "etl",
    { source: "s3://bucket/data.csv" },
    { edges: [uuidB, uuidC] }
  ),
  
  // Job B (depends on A)
  BaseDAGBuilder.buildJobNode(
    uuidB,
    "Transform Batch 1",
    "etl",
    { batch: 1 },
    { edges: [uuidD] }
  ),
  
  // Job C (depends on A, parallel with B)
  BaseDAGBuilder.buildJobNode(
    uuidC,
    "Transform Batch 2",
    "etl",
    { batch: 2 },
    { edges: [uuidD] }
  ),
  
  // Job D (depends on B and C)
  BaseDAGBuilder.buildJobNode(
    uuidD,
    "Merge Results",
    "etl",
    { operation: "merge" },
    { edges: [] }
  )
];

// Validate and submit
const validation = BaseDAGBuilder.validateDAG(dag);
if (!validation.isValid) {
  throw new Error(`Validation failed: ${validation.error}`);
}

await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dag)
});
```

---

## Method 3: Python Client Library

### Installation

No installation required - pure Python library. Just import:

```python
import sys
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))

from dag_builders.contact_import import create_contact_import_dag
```

### Example 1: Contact Import DAG

```python
from dag_builders.contact_import import create_contact_import_dag
import requests

# Basic usage
dag = create_contact_import_dag(
    s3_bucket="my-uploads-bucket",
    s3_key="uploads/contacts_2024.csv"
)

# With options
dag = create_contact_import_dag(
    s3_bucket="my-uploads-bucket",
    s3_key="uploads/contacts_2024.csv",
    workflow_id="custom-workflow-id",
    retry_count=5,
    retry_interval=10,
    custom_vars={
        "source": "manual_upload",
        "user_id": "user-123"
    }
)

# Submit to API
response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={
        "Content-Type": "application/json",
        "X-API-Key": "your-api-key"
    },
    json=dag
)

print(response.json())
```

### Example 2: Email Export DAG

```python
from dag_builders.email_export import create_email_export_dag

dag = create_email_export_dag(
    s3_bucket="my-exports-bucket",
    s3_key="exports/contacts.csv",
    output_s3_key="exports/enriched/contacts_enriched.csv",
    email_api_url="https://api.emailfinder.com/v1",
    email_api_key="api-key-123",
    batch_size=100,
    retry_count=5,
    retry_interval=10
)

response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dag
)
```

### Example 3: Custom DAG Using BaseDAGBuilder

```python
from dag_builders.base import BaseDAGBuilder

# Generate workflow ID
workflow_id = BaseDAGBuilder.generate_workflow_id("my-workflow")

# Build custom job nodes
dag = [
    BaseDAGBuilder.build_job_node(
        BaseDAGBuilder.generate_uuid("extract"),
        "Extract Data",
        "etl",
        {
            "source": "s3://bucket/data.csv",
            "format": "csv",
            "custom_vars": {
                "date": "2024-01-15",
                "region": "us-east-1"
            }
        },
        edges=[BaseDAGBuilder.generate_uuid("transform")],
        retry_count=3,
        retry_interval=5
    ),
    BaseDAGBuilder.build_job_node(
        BaseDAGBuilder.generate_uuid("transform"),
        "Transform Data",
        "etl",
        {"transformations": ["clean", "normalize"]},
        edges=[],
        retry_count=3,
        retry_interval=5
    )
]

# Validate before submission
is_valid, error = BaseDAGBuilder.validate_dag(dag)
if not is_valid:
    raise ValueError(f"DAG validation failed: {error}")

# Submit to API
response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dag
)
```

### Example 4: Complex DAG with Multiple Dependencies

```python
from dag_builders.base import BaseDAGBuilder

# Generate UUIDs
uuid_a = BaseDAGBuilder.generate_uuid("job-a")
uuid_b = BaseDAGBuilder.generate_uuid("job-b")
uuid_c = BaseDAGBuilder.generate_uuid("job-c")
uuid_d = BaseDAGBuilder.generate_uuid("job-d")

dag = [
    # Job A (source, no dependencies)
    BaseDAGBuilder.build_job_node(
        uuid_a,
        "Extract Data",
        "etl",
        {"source": "s3://bucket/data.csv"},
        edges=[uuid_b, uuid_c]
    ),
    
    # Job B (depends on A)
    BaseDAGBuilder.build_job_node(
        uuid_b,
        "Transform Batch 1",
        "etl",
        {"batch": 1},
        edges=[uuid_d]
    ),
    
    # Job C (depends on A, parallel with B)
    BaseDAGBuilder.build_job_node(
        uuid_c,
        "Transform Batch 2",
        "etl",
        {"batch": 2},
        edges=[uuid_d]
    ),
    
    # Job D (depends on B and C)
    BaseDAGBuilder.build_job_node(
        uuid_d,
        "Merge Results",
        "etl",
        {"operation": "merge"},
        edges=[]
    )
]

# Validate and submit
is_valid, error = BaseDAGBuilder.validate_dag(dag)
if not is_valid:
    raise ValueError(f"Validation failed: {error}")

response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dag
)
```

---

## DAG Structure Deep Dive

### Job Node Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `uuid` | string | ✅ Yes | Unique job identifier |
| `job_title` | string | ✅ Yes | Human-readable job name |
| `job_type` | string | ✅ Yes | Type of job (e.g., "etl", "insert_csv_file") |
| `data` | object | ✅ Yes | Job payload (stored as JSONB) |
| `edges` | array | ✅ Yes | Array of target UUIDs (dependencies) |

### Job Node Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `retry_count` | number | 0 | Number of retry attempts |
| `retry_interval` | number | 30 | Minutes between retries |
| `run_after` | string | now | ISO 8601 datetime (earliest execution time) |

### Data Field Structure

The `data` field accepts any JSON structure. Common patterns:

```json
{
  "data": {
    // Workflow metadata
    "workflow_uuid": "workflow-123",
    "workflow_type": "contact_import",
    "created_at": "2024-01-15T10:00:00Z",
    
    // Job-specific data
    "s3_bucket": "my-bucket",
    "s3_key": "uploads/data.csv",
    
    // Custom variables (for filtering/querying)
    "custom_vars": {
      "user_id": "user-123",
      "campaign_id": "campaign-456",
      "source": "api_upload"
    },
    
    // Configuration
    "batch_size": 100,
    "format": "csv",
    "config": {
      "parallelism": 5,
      "timeout": 300
    }
  }
}
```

---

## Dependency Patterns

### Pattern 1: Sequential Pipeline

**Structure**: A → B → C → D

```json
[
  {"uuid": "a", "edges": ["b"]},
  {"uuid": "b", "edges": ["c"]},
  {"uuid": "c", "edges": ["d"]},
  {"uuid": "d", "edges": []}
]
```

**Execution**: Jobs execute one after another sequentially.

### Pattern 2: Parallel Fan-out

**Structure**: A → [B, C, D] → E

```json
[
  {"uuid": "a", "edges": ["b", "c", "d"]},
  {"uuid": "b", "edges": ["e"]},
  {"uuid": "c", "edges": ["e"]},
  {"uuid": "d", "edges": ["e"]},
  {"uuid": "e", "edges": []}
]
```

**Execution**: After A completes, B, C, D execute in parallel, then E executes.

### Pattern 3: Mixed Dependencies

**Structure**: A → [B, C], B → D, C → D

```json
[
  {"uuid": "a", "edges": ["b", "c"]},
  {"uuid": "b", "edges": ["d"]},
  {"uuid": "c", "edges": ["d"]},
  {"uuid": "d", "edges": []}
]
```

**Execution**: A → B and C (parallel) → D (waits for both B and C).

### Pattern 4: Multiple Independent Paths

**Structure**: Two independent paths: A → B and C → D

```json
[
  {"uuid": "a", "edges": ["b"]},
  {"uuid": "b", "edges": []},
  {"uuid": "c", "edges": ["d"]},
  {"uuid": "d", "edges": []}
]
```

**Execution**: Both paths execute independently in parallel.

---

## DAG Validation

### Automatic Validation

The system automatically validates DAGs on creation:

1. **Cycle Detection**: Rejects DAGs with circular dependencies
2. **In-degree Calculation**: Calculates and stores in-degree for each job
3. **Topological Sort**: Verifies DAG can be processed (Kahn's Algorithm)
4. **UUID Validation**: Checks for duplicate UUIDs
5. **Edge Validation**: Ensures all edge targets exist in DAG

### Invalid DAG Example (Cycle)

```json
[
  {"uuid": "job-a", "edges": ["job-b"]},
  {"uuid": "job-b", "edges": ["job-c"]},
  {"uuid": "job-c", "edges": ["job-a"]}  // ❌ Creates cycle: a → b → c → a
]
```

**Error**: `400 Bad Request` - "validation failed: directed acyclic graph contains cycles or invalid structure"

### Valid DAG Example

```json
[
  {"uuid": "job-a", "edges": ["job-b"]},
  {"uuid": "job-b", "edges": ["job-c"]},
  {"uuid": "job-c", "edges": []}  // ✅ No cycles
]
```

### Client-Side Validation

**TypeScript:**
```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

const dag = createMyDAG(...);
const validation = BaseDAGBuilder.validateDAG(dag);

if (!validation.isValid) {
  console.error(`Validation failed: ${validation.error}`);
  // Don't submit
}
```

**Python:**
```python
from dag_builders.base import BaseDAGBuilder

dag = create_my_dag(...)
is_valid, error = BaseDAGBuilder.validate_dag(dag)

if not is_valid:
    print(f"Validation failed: {error}")
    # Don't submit
```

---

## Complete Examples

### Example 1: ETL Pipeline (Extract → Transform → Load)

**TypeScript:**
```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

const extractId = BaseDAGBuilder.generateUUID("extract");
const transformId = BaseDAGBuilder.generateUUID("transform");
const loadId = BaseDAGBuilder.generateUUID("load");

const dag = [
  BaseDAGBuilder.buildJobNode(
    extractId,
    "Extract Data from S3",
    "etl",
    {
      source: "s3://bucket/data.csv",
      format: "csv",
      custom_vars: { date: "2024-01-15" }
    },
    { edges: [transformId], retryCount: 3 }
  ),
  BaseDAGBuilder.buildJobNode(
    transformId,
    "Transform Data",
    "etl",
    {
      transformations: ["clean", "normalize", "validate"],
      output_format: "parquet"
    },
    { edges: [loadId], retryCount: 3 }
  ),
  BaseDAGBuilder.buildJobNode(
    loadId,
    "Load to Warehouse",
    "etl",
    {
      destination: "warehouse",
      table: "daily_data",
      mode: "append"
    },
    { edges: [], retryCount: 3 }
  )
];

// Submit
await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dag)
});
```

**Python:**
```python
from dag_builders.base import BaseDAGBuilder
import requests

extract_id = BaseDAGBuilder.generate_uuid("extract")
transform_id = BaseDAGBuilder.generate_uuid("transform")
load_id = BaseDAGBuilder.generate_uuid("load")

dag = [
    BaseDAGBuilder.build_job_node(
        extract_id,
        "Extract Data from S3",
        "etl",
        {
            "source": "s3://bucket/data.csv",
            "format": "csv",
            "custom_vars": {"date": "2024-01-15"}
        },
        edges=[transform_id],
        retry_count=3
    ),
    BaseDAGBuilder.build_job_node(
        transform_id,
        "Transform Data",
        "etl",
        {
            "transformations": ["clean", "normalize", "validate"],
            "output_format": "parquet"
        },
        edges=[load_id],
        retry_count=3
    ),
    BaseDAGBuilder.build_job_node(
        load_id,
        "Load to Warehouse",
        "etl",
        {
            "destination": "warehouse",
            "table": "daily_data",
            "mode": "append"
        },
        edges=[],
        retry_count=3
    )
]

response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dag
)
```

### Example 2: Contact Import with Email Enrichment

**TypeScript:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";
import { createEmailExportDAG } from "./src/builders/EmailExportDAGBuilder";
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

// Step 1: Import contacts
const importDag = createContactImportDAG(
  "my-bucket",
  "uploads/contacts.csv",
  { workflowId: "workflow-123" }
);

const importJobId = importDag[0].uuid;

// Step 2: Enrich with emails (depends on import)
const exportDag = createEmailExportDAG(
  "my-bucket",
  "exports/contacts.csv",
  {
    workflowId: "workflow-123",
    emailApiUrl: "https://api.emailfinder.com/v1",
    emailApiKey: "api-key-123"
  }
);

// Add dependency
exportDag[0].edges = [importJobId];

// Combine DAGs
const completeDag = [...importDag, ...exportDag];

// Submit
await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(completeDag)
});
```

### Example 3: Batch Processing with Merge

**Python:**
```python
from dag_builders.base import BaseDAGBuilder
import requests

# Generate UUIDs
split_id = BaseDAGBuilder.generate_uuid("split")
merge_id = BaseDAGBuilder.generate_uuid("merge")
batch_ids = [BaseDAGBuilder.generate_uuid(f"batch-{i}") for i in range(3)]

# Build DAG
dag = [
    # Split job (source)
    BaseDAGBuilder.build_job_node(
        split_id,
        "Split Data into Batches",
        "etl",
        {"operation": "split", "batch_count": 3},
        edges=batch_ids,
        retry_count=3
    )
]

# Add batch processing jobs
for i, batch_id in enumerate(batch_ids):
    dag.append(
        BaseDAGBuilder.build_job_node(
            batch_id,
            f"Process Batch {i+1}",
            "etl",
            {"batch": i+1, "batch_size": 1000},
            edges=[merge_id],
            retry_count=3
        )
    )

# Add merge job
dag.append(
    BaseDAGBuilder.build_job_node(
        merge_id,
        "Merge Results",
        "etl",
        {"operation": "merge"},
        edges=[],
        retry_count=3
    )
)

# Validate and submit
is_valid, error = BaseDAGBuilder.validate_dag(dag)
if not is_valid:
    raise ValueError(f"Validation failed: {error}")

response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dag
)
```

---

## Task Breakdown

### Phase 1: Understanding DAG Structure ✅

- [x] Understand what a DAG is (Directed Acyclic Graph)
- [x] Learn job node format (uuid, job_title, job_type, data, edges)
- [x] Understand dependency relationships (edges array)
- [x] Learn degree concept (in-degree = number of dependencies)

### Phase 2: Manual JSON Creation ✅

- [x] Learn API endpoint structure (`POST /jobs/bulk-insert/complete-graph`)
- [x] Understand request headers (Content-Type, X-API-Key)
- [x] Practice creating simple sequential DAGs
- [x] Practice creating parallel fan-out DAGs
- [x] Practice creating mixed dependency DAGs

### Phase 3: TypeScript Client Library ✅

- [x] Study BaseDAGBuilder class (UUID generation, validation)
- [x] Study ContactImportDAGBuilder
- [x] Study EmailExportDAGBuilder
- [x] Study UnifiedDAGBuilder
- [x] Practice creating DAGs with TypeScript
- [x] Learn validation before submission

### Phase 4: Python Client Library ✅

- [x] Study base.py (BaseDAGBuilder class)
- [x] Study contact_import.py
- [x] Study email_export.py
- [x] Study unified.py
- [x] Practice creating DAGs with Python
- [x] Learn validation before submission

### Phase 5: DAG Validation ✅

- [x] Understand cycle detection (Kahn's Algorithm)
- [x] Learn in-degree calculation process
- [x] Study validation errors and how to fix them
- [x] Practice client-side validation

### Phase 6: Advanced Patterns ✅

- [x] Create complex multi-level DAGs
- [x] Combine multiple builder functions
- [x] Use custom variables for metadata
- [x] Configure retry settings appropriately
- [x] Monitor DAG progress

---

## Best Practices

1. **Always Validate Before Submission**: Use client-side validation to catch errors early
2. **Use Descriptive UUIDs**: Include date, type, and purpose (e.g., `etl-2024-01-15-extract`)
3. **Store Configuration in `data`**: Use the `data` field for job-specific configuration
4. **Set Appropriate Retry Counts**: Balance between resilience and avoiding infinite retries
5. **Use Custom Variables**: Add metadata for filtering/querying later
6. **Group Related Jobs**: Use common prefixes for easier querying
7. **Monitor DAG Progress**: Use `/dag/status` endpoint to track execution

---

## Common Mistakes to Avoid

1. ❌ **Circular Dependencies**: Don't create cycles (A → B → C → A)
2. ❌ **Missing Edge Targets**: Ensure all edge targets exist in DAG
3. ❌ **Duplicate UUIDs**: Each job must have a unique UUID
4. ❌ **Missing Required Fields**: Always include uuid, job_title, job_type, data
5. ❌ **Invalid JSON**: Ensure proper JSON formatting

---

## Additional Resources

- [API Documentation](./API.md) - Complete API reference
- [DAG Guide](./DAG_GUIDE.md) - Understanding DAGs and dependencies
- [DAG Builders Guide](./DAG_BUILDERS_GUIDE.md) - Detailed builder documentation
- [TypeScript Examples](../client-libs/typescript/examples/) - Working examples
- [Python Examples](../client-libs/python/examples/) - Working examples

---

## Summary

Creating DAGs in the Job Scheduler system is straightforward:

1. **Choose your method**: Manual JSON, TypeScript, or Python
2. **Build your DAG**: Define jobs and dependencies
3. **Validate**: Check for cycles and errors
4. **Submit**: POST to `/jobs/bulk-insert/complete-graph`
5. **Monitor**: Track progress using status endpoints

The system automatically handles dependency resolution, cycle detection, and parallel execution. You just need to define the structure!
