# DAG Creation and Management Guide

Comprehensive guide for creating and managing Directed Acyclic Graphs (DAGs) in the Job Scheduler.

## Overview

A DAG (Directed Acyclic Graph) is a collection of jobs with dependencies where each job can depend on one or more other jobs. The scheduler automatically resolves dependencies and executes jobs in the correct order.

## DAG Structure

### Visual Representation

```
┌─────────────────────────────────────┐
│          DAG Example                │
├─────────────────────────────────────┤
│                                     │
│  Job A (degree: 0)                  │
│    │                                │
│    ├──────────┬─────────────────┐   │
│    │          │                 │   │
│    ▼          ▼                 ▼   │
│  Job B     Job C             Job D  │
│  (degree: 1) (degree: 1)  (degree: 1)
│    │          │                 │   │
│    └──────┬───┘                 │   │
│           │                     │   │
│           ▼                     ▼   │
│         Job E                 Job F │
│      (degree: 2)          (degree: 2)
│                                     │
└─────────────────────────────────────┘
```

### Key Concepts

- **Node (Job)**: A task to be executed
- **Edge (Dependency)**: Relationship where source job must complete before target job
- **Degree (In-degree)**: Number of dependencies a job has (calculated automatically)
- **Topological Sort**: Algorithm that determines execution order

## Creating a DAG

### Basic Example

```json
POST /jobs/bulk-insert/complete-graph

[
  {
    "uuid": "job-a",
    "job_title": "Extract Data",
    "job_type": "etl",
    "data": {"source": "s3://bucket/data"},
    "retry_count": 3,
    "edges": ["job-b"]
  },
  {
    "uuid": "job-b",
    "job_title": "Transform Data",
    "job_type": "etl",
    "data": {"transform": "aggregate"},
    "retry_count": 3,
    "edges": []
  }
]
```

**Result**: `job-a` must complete before `job-b` can execute.

### Complex DAG Example

```json
[
  {
    "uuid": "extract",
    "job_title": "Extract Data",
    "job_type": "etl",
    "data": {"source": "s3://bucket/data"},
    "edges": ["transform"]
  },
  {
    "uuid": "transform",
    "job_title": "Transform Data",
    "job_type": "etl",
    "data": {"transformations": ["clean", "normalize"]},
    "edges": ["load", "report"]
  },
  {
    "uuid": "load",
    "job_title": "Load to Warehouse",
    "job_type": "etl",
    "data": {"destination": "warehouse"},
    "edges": []
  },
  {
    "uuid": "report",
    "job_title": "Generate Report",
    "job_type": "report",
    "data": {"format": "pdf"},
    "edges": []
  }
]
```

**Execution Flow**:
1. `extract` runs first (degree: 0)
2. `extract` completes → `transform` becomes ready (degree: 1 → 0)
3. `transform` runs
4. `transform` completes → `load` and `report` become ready (both degree: 1 → 0)
5. `load` and `report` run in parallel (both degree: 0)

## Dynamic DAG Creation

### Using UUID Prefixes

```python
import requests
from datetime import datetime

def create_dynamic_etl_dag(date_str, bucket, config):
    """Create ETL DAG dynamically with custom variables"""
    
    base_uuid_prefix = f"etl-{date_str}"
    
    dag = [
        {
            "uuid": f"{base_uuid_prefix}-extract",
            "job_title": f"Extract Data - {date_str}",
            "job_type": "etl",
            "data": {
                "source_bucket": bucket,
                "source_key": f"data/{date_str}/input.csv",
                "format": config.get("format", "csv"),
                "custom_vars": {
                    "date": date_str,
                    "region": config.get("region", "us-east-1")
                }
            },
            "retry_count": config.get("retry_count", 3),
            "retry_interval": config.get("retry_interval", 5),
            "edges": [f"{base_uuid_prefix}-transform"]
        },
        {
            "uuid": f"{base_uuid_prefix}-transform",
            "job_title": f"Transform Data - {date_str}",
            "job_type": "etl",
            "data": {
                "transformations": config.get("transformations", []),
                "output_format": config.get("output_format", "parquet")
            },
            "retry_count": config.get("retry_count", 3),
            "retry_interval": config.get("retry_interval", 5),
            "edges": [f"{base_uuid_prefix}-load"]
        },
        {
            "uuid": f"{base_uuid_prefix}-load",
            "job_title": f"Load to Warehouse - {date_str}",
            "job_type": "etl",
            "data": {
                "destination": config.get("warehouse", "warehouse"),
                "table": config.get("table", "daily_data")
            },
            "edges": []
        }
    ]
    
    return dag

# Usage
date_str = "2024-01-15"
config = {
    "format": "csv",
    "transformations": ["clean", "normalize"],
    "output_format": "parquet",
    "warehouse": "data_warehouse",
    "table": "daily_processed"
}

dag = create_dynamic_etl_dag(date_str, "my-bucket", config)

response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={
        "Content-Type": "application/json",
        "X-API-Key": "your-api-key"
    },
    json=dag
)
```

### Using Templates

Create reusable DAG templates for common workflows:

```python
TEMPLATE_ETL_PIPELINE = {
    "nodes": [
        {"uuid_template": "extract-{date}", "edges_template": ["transform-{date}"]},
        {"uuid_template": "transform-{date}", "edges_template": ["load-{date}"]},
        {"uuid_template": "load-{date}", "edges_template": []}
    ]
}

def process_template(template, variables):
    """Process DAG template with variables"""
    import re
    
    def replace_vars(text):
        if isinstance(text, str):
            for key, value in variables.items():
                text = text.replace(f"{{{key}}}", str(value))
            return text
        elif isinstance(text, dict):
            return {k: replace_vars(v) for k, v in text.items()}
        elif isinstance(text, list):
            return [replace_vars(item) for item in text]
        return text
    
    dag = replace_vars(template["nodes"])
    return dag
```

## DAG Validation

### Automatic Validation

The system automatically validates DAGs on creation:

1. **Cycle Detection**: Rejects DAGs with circular dependencies
2. **In-degree Calculation**: Calculates and stores in-degree for each job
3. **Topological Sort**: Verifies DAG can be processed

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

## Job Data (Custom Variables)

### Storing Custom Data

The `data` field accepts any JSON structure:

```json
{
  "uuid": "job-custom",
  "job_title": "Custom Job",
  "job_type": "custom",
  "data": {
    "source": "s3://bucket/data",
    "destination": "s3://bucket/output",
    "transformations": ["clean", "normalize", "validate"],
    "custom_vars": {
      "date": "2024-01-15",
      "region": "us-east-1",
      "environment": "production",
      "quality_threshold": 0.95
    },
    "config": {
      "batch_size": 1000,
      "parallelism": 5
    }
  },
  "edges": []
}
```

### Accessing Data in Processors

Processors can access job data:

```go
func (p *CustomProcessor) Process(ctx context.Context, job *models.ModelJobs) error {
    var jobData map[string]interface{}
    if err := json.Unmarshal(job.Data, &jobData); err != nil {
        return err
    }
    
    source := jobData["source"].(string)
    customVars := jobData["custom_vars"].(map[string]interface{})
    
    // Use data...
}
```

## Checking DAG Status

### Get All Jobs in DAG

```bash
GET /jobs/dag/status?uuids=job-a&uuids=job-b&uuids=job-c
```

### Monitor DAG Progress

```python
import requests
import time

def monitor_dag_progress(api_url, api_key, job_uuids, poll_interval=5):
    """Monitor DAG progress with polling"""
    
    while True:
        response = requests.get(
            f"{api_url}/jobs/dag/status",
            params={"uuids": job_uuids},
            headers={"X-API-Key": api_key}
        )
        
        status = response.json()
        
        print(f"DAG Status: {status['dag_status']}")
        print(f"Progress: {status['progress']['percentage']}%")
        print(f"Status Breakdown: {status['status_breakdown']}")
        
        if status['dag_status'] in ['completed', 'failed']:
            break
        
        time.sleep(poll_interval)

# Usage
monitor_dag_progress(
    api_url="http://localhost:8000",
    api_key="your-api-key",
    job_uuids=["job-a", "job-b", "job-c"]
)
```

## DAG Patterns

### Parallel Fan-out

```json
[
  {"uuid": "split", "edges": ["job-1", "job-2", "job-3"]},
  {"uuid": "job-1", "edges": ["merge"]},
  {"uuid": "job-2", "edges": ["merge"]},
  {"uuid": "job-3", "edges": ["merge"]},
  {"uuid": "merge", "edges": []}
]
```

All three jobs (`job-1`, `job-2`, `job-3`) execute in parallel after `split` completes.

### Sequential Pipeline

```json
[
  {"uuid": "step1", "edges": ["step2"]},
  {"uuid": "step2", "edges": ["step3"]},
  {"uuid": "step3", "edges": ["step4"]},
  {"uuid": "step4", "edges": []}
]
```

Jobs execute sequentially: step1 → step2 → step3 → step4.

### Mixed Dependencies

```json
[
  {"uuid": "a", "edges": ["b", "c"]},
  {"uuid": "b", "edges": ["d"]},
  {"uuid": "c", "edges": ["d"]},
  {"uuid": "d", "edges": []}
]
```

Execution: `a` → `b` and `c` (parallel) → `d`.

## Best Practices

1. **Use Descriptive UUIDs**: Include date, type, and purpose (e.g., `etl-2024-01-15-extract`)
2. **Keep DAGs Acyclic**: Avoid circular dependencies
3. **Group Related Jobs**: Use common prefixes for easier querying
4. **Store Configuration in `data`**: Use the `data` field for job-specific configuration
5. **Set Appropriate Retry Counts**: Balance between resilience and avoiding infinite retries
6. **Monitor DAG Progress**: Use `/dag/status` and `/dag/progress` endpoints

## Common Patterns

### ETL Pipeline

```
Extract → Transform → Load
```

### CI/CD Pipeline

```
Build → Test → Deploy → Notify
```

### Data Processing

```
Ingest → Validate → Transform → Aggregate → Export
```

### Report Generation

```
Fetch Data → Calculate Metrics → Generate Report → Send Notification
```

---

## See Also

- [API Documentation](./API.md) - Complete API reference
- [Configuration Guide](./CONFIGURATION.md) - Environment variables
- [README.md](../README.md) - Project overview
