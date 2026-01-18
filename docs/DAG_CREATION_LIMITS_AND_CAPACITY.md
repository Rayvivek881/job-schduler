# Complete Guide: How Many DAGs Can We Create?

This comprehensive guide explains the limits, constraints, and strategies for creating DAGs in the Job Scheduler system.

---

## 📚 Table of Contents

1. [Key Limits Overview](#key-limits-overview)
2. [Single DAG Limits](#single-dag-limits)
3. [API Rate Limits](#api-rate-limits)
4. [System Capacity](#system-capacity)
5. [Strategies for Large-Scale DAG Creation](#strategies-for-large-scale-dag-creation)
6. [Batching and Chunking Patterns](#batching-and-chunking-patterns)
7. [Scaling Strategies](#scaling-strategies)
8. [Best Practices](#best-practices)
9. [Task Breakdown](#task-breakdown)

---

## Key Limits Overview

### Summary Table

| Limit Type | Value | Location | Configurable |
|------------|-------|----------|--------------|
| **Max Nodes per Request** | 5,000 | `constants.MaxNodesPerRequest` | ✅ Yes |
| **API Rate Limit** | 1,000-5,000 req/min | `MAX_REQUESTS_PER_MINUTE` | ✅ Yes |
| **Query Pagination Limit** | 100 per page | `constants.MaxPageLimit` | ✅ Yes |
| **Scheduler Batch Size** | 100 jobs per tick | `constants.MaxPageLimit` | ✅ Yes |
| **No Limit on DAG Count** | Unlimited | - | - |

### Key Insights

- ✅ **Single DAG**: Up to 5,000 jobs (nodes)
- ✅ **Multiple DAGs**: Unlimited (limited by rate limiting)
- ✅ **Rate Limiting**: 1,000-5,000 requests per minute
- ✅ **Large Scenarios**: Must chunk/batch DAGs

---

## Single DAG Limits

### Maximum Nodes per Request: 5,000

**Location**: `constants/jobs.go:13`
```go
MaxNodesPerRequest = 5000
```

**What it means:**
- **Single DAG request** can contain maximum **5,000 job nodes**
- This includes all jobs in the DAG and their dependencies
- Edge relationships are also included in the request

**Example:**
```json
// ✅ Valid: 5,000 jobs in single DAG
[
  {"uuid": "job-1", ...},
  {"uuid": "job-2", ...},
  ...
  {"uuid": "job-5000", ...}
]

// ❌ Invalid: 5,001 jobs exceeds limit
[
  {"uuid": "job-1", ...},
  ...
  {"uuid": "job-5001", ...}  // ERROR: Exceeds MaxNodesPerRequest
]
```

**Validation:**
- **Client-side**: TypeScript/Python builders validate before submission
- **Server-side**: API validates in `server/helper/requests.go:69`

**Error Response:**
```json
{
  "error": "validation failed: directed acyclic graph contains cycles or invalid structure",
  "success": false
}
```

### Why This Limit?

1. **Memory Protection**: Prevents server memory exhaustion
2. **Processing Time**: Large DAGs take time to validate and insert
3. **Database Performance**: Bulk inserts are optimized for reasonable sizes
4. **Network Payload**: Large JSON payloads can cause timeouts

---

## API Rate Limits

### Rate Limit: 1,000-5,000 Requests per Minute

**Location**: `conf/viper.go` → `MAX_REQUESTS_PER_MINUTE`

**Default**: 1,000 requests/minute
**Production**: 5,000 requests/minute (configurable)

**What it means:**
- You can make **1,000-5,000 API requests per minute**
- Each DAG creation counts as **1 request**
- Rate limiting uses **Token Bucket Algorithm**

### Rate Limiter Algorithm

**Implementation**: `middleware/rateMiddleware.go`

```go
// Token Bucket Algorithm
maxLimit = 1000  // Max tokens (configurable)
fillingRate = 1000 / 60  // Tokens per second (~16.67)

// Background refiller (runs every second)
go func() {
    ticker := time.NewTicker(time.Second)
    for range ticker.C {
        tokens = min(maxLimit, tokens + fillingRate)
    }
}()

// Request handler
func RateLimiter() gin.HandlerFunc {
    return func(c *gin.Context) {
        if tokens <= 0 {
            c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
            return
        }
        tokens--  // Consume one token
        c.Next()
    }
}
```

**Characteristics:**
- ✅ Smooth rate limiting (continuous refill, not burst)
- ✅ Thread-safe (mutex protection)
- ✅ Configurable via `MAX_REQUESTS_PER_MINUTE` environment variable

### Rate Limit Calculations

**Scenario 1: Small DAGs (10 nodes each)**
```
Rate Limit: 1,000 requests/minute
DAG Size: 10 nodes each
Max DAGs: 1,000 DAGs/minute
Max Jobs: 1,000 DAGs × 10 nodes = 10,000 jobs/minute
```

**Scenario 2: Medium DAGs (100 nodes each)**
```
Rate Limit: 1,000 requests/minute
DAG Size: 100 nodes each
Max DAGs: 1,000 DAGs/minute
Max Jobs: 1,000 DAGs × 100 nodes = 100,000 jobs/minute
```

**Scenario 3: Large DAGs (5,000 nodes each)**
```
Rate Limit: 1,000 requests/minute
DAG Size: 5,000 nodes each (maximum)
Max DAGs: 1,000 DAGs/minute
Max Jobs: 1,000 DAGs × 5,000 nodes = 5,000,000 jobs/minute
```

**Scenario 4: Production Rate Limit (5,000 req/min)**
```
Rate Limit: 5,000 requests/minute
DAG Size: 5,000 nodes each
Max DAGs: 5,000 DAGs/minute
Max Jobs: 5,000 DAGs × 5,000 nodes = 25,000,000 jobs/minute
```

### Rate Limit Error

**HTTP Status**: `429 Too Many Requests`

**Response:**
```json
{
  "error": "rate limit exceeded: too many requests, please retry after some time",
  "success": false
}
```

**Solution**: Wait for token refill or increase `MAX_REQUESTS_PER_MINUTE`

---

## System Capacity

### Unlimited DAG Creation

**Key Point**: There is **no explicit limit on the total number of DAGs** you can create.

**Constraints are:**
1. **Per-request limit**: 5,000 nodes per DAG
2. **Rate limiting**: 1,000-5,000 requests per minute
3. **System resources**: Database, Kafka, memory, CPU

### Theoretical Maximums

**Per Minute:**
```
Max DAGs = Rate Limit (1,000-5,000 req/min)
Max Jobs = Max DAGs × 5,000 nodes = 5,000,000-25,000,000 jobs/min
```

**Per Hour:**
```
Max DAGs = Rate Limit × 60 = 60,000-300,000 DAGs/hour
Max Jobs = Max DAGs × 5,000 nodes = 300,000,000-1,500,000,000 jobs/hour
```

**Per Day:**
```
Max DAGs = Rate Limit × 1,440 = 1,440,000-7,200,000 DAGs/day
Max Jobs = Max DAGs × 5,000 nodes = 7,200,000,000-36,000,000,000 jobs/day
```

**Note**: These are theoretical maximums. Real-world capacity depends on:
- Database performance
- Kafka throughput
- Network bandwidth
- System resources (CPU, memory, disk)

### Practical Considerations

**Database Capacity:**
- PostgreSQL can handle millions of rows
- Indexes optimize query performance
- Connection pooling (40 max connections)

**Kafka Capacity:**
- Partitions determine parallelism
- Consumer groups distribute load
- Message throughput: 100+ jobs/second per consumer

**Memory/CPU:**
- Each DAG validation uses CPU
- Large JSON payloads use memory
- Multiple concurrent requests increase load

---

## Strategies for Large-Scale DAG Creation

### Strategy 1: Chunking Large DAGs

If you need more than 5,000 nodes, split into multiple DAGs.

**Python Example:**
```python
from dag_builders.contact_import import create_contact_import_dag
import requests

def create_large_batch_dags(file_list: list, chunk_size: int = 5000):
    """Create multiple DAGs by chunking file list"""
    
    all_dags = []
    
    # Split files into chunks of chunk_size
    for i in range(0, len(file_list), chunk_size):
        chunk = file_list[i:i+chunk_size]
        chunk_dags = []
        
        for file in chunk:
            dag = create_contact_import_dag(
                s3_bucket="my-bucket",
                s3_key=file,
                custom_vars={"batch_id": f"batch-{i//chunk_size + 1}"}
            )
            chunk_dags.extend(dag)
        
        all_dags.append(chunk_dags)
    
    return all_dags

# Usage: 50,000 files
file_list = [f"uploads/file_{i}.csv" for i in range(50000)]

# Create DAGs in chunks of 5,000
dag_chunks = create_large_batch_dags(file_list, chunk_size=5000)
# Result: 10 DAGs, each with 5,000 jobs

# Submit each chunk sequentially
api_url = "http://localhost:8000/jobs/bulk-insert/complete-graph"
api_key = "your-api-key"

for chunk_dags in dag_chunks:
    response = requests.post(
        api_url,
        headers={"X-API-Key": api_key},
        json=chunk_dags
    )
    print(f"Submitted chunk: {len(chunk_dags)} jobs")
```

**TypeScript Example:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

function createLargeBatchDAGs(
  fileList: string[],
  chunkSize: number = 5000
): any[][] {
  const allDAGs: any[][] = [];
  
  // Split files into chunks
  for (let i = 0; i < fileList.length; i += chunkSize) {
    const chunk = fileList.slice(i, i + chunkSize);
    const chunkDAGs: any[] = [];
    
    chunk.forEach((file) => {
      const dag = createContactImportDAG("my-bucket", file, {
        customVars: { batch_id: `batch-${Math.floor(i / chunkSize) + 1}` }
      });
      chunkDAGs.push(...dag);
    });
    
    allDAGs.push(chunkDAGs);
  }
  
  return allDAGs;
}

// Usage: 50,000 files
const fileList = Array.from({ length: 50000 }, (_, i) => `uploads/file_${i}.csv`);

// Create DAGs in chunks of 5,000
const dagChunks = createLargeBatchDAGs(fileList, 5000);
// Result: 10 DAGs, each with 5,000 jobs

// Submit each chunk sequentially
const apiUrl = "http://localhost:8000/jobs/bulk-insert/complete-graph";
const apiKey = "your-api-key";

for (const chunkDAGs of dagChunks) {
  await fetch(apiUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": apiKey
    },
    body: JSON.stringify(chunkDAGs)
  });
  console.log(`Submitted chunk: ${chunkDAGs.length} jobs`);
}
```

### Strategy 2: Parallel DAG Submission

Submit multiple DAGs in parallel (respecting rate limits).

**Python Example:**
```python
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed
from dag_builders.contact_import import create_contact_import_dag

def submit_dag(dag, api_url, api_key):
    """Submit single DAG to API"""
    response = requests.post(
        api_url,
        headers={"X-API-Key": api_key},
        json=dag
    )
    return response.status_code == 200

def submit_dags_parallel(dags: list, max_workers: int = 10):
    """Submit multiple DAGs in parallel"""
    
    api_url = "http://localhost:8000/jobs/bulk-insert/complete-graph"
    api_key = "your-api-key"
    
    results = []
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = {
            executor.submit(submit_dag, dag, api_url, api_key): idx
            for idx, dag in enumerate(dags)
        }
        
        for future in as_completed(futures):
            idx = futures[future]
            try:
                success = future.result()
                results.append({"idx": idx, "success": success})
            except Exception as e:
                results.append({"idx": idx, "success": False, "error": str(e)})
    
    return results

# Usage
file_list = [f"uploads/file_{i}.csv" for i in range(1000)]
dags = [
    create_contact_import_dag("my-bucket", file)[0]
    for file in file_list
]

# Submit 1000 DAGs in parallel (10 workers)
results = submit_dags_parallel(dags, max_workers=10)
successful = sum(1 for r in results if r["success"])
print(f"Submitted {successful}/{len(dags)} DAGs successfully")
```

**TypeScript Example:**
```typescript
async function submitDAG(dag: any[], apiUrl: string, apiKey: string): Promise<boolean> {
  const response = await fetch(apiUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": apiKey
    },
    body: JSON.stringify(dag)
  });
  return response.ok;
}

async function submitDAGsParallel(
  dags: any[][],
  maxConcurrent: number = 10
): Promise<Array<{ idx: number; success: boolean }>> {
  const apiUrl = "http://localhost:8000/jobs/bulk-insert/complete-graph";
  const apiKey = "your-api-key";
  
  const results: Array<{ idx: number; success: boolean }> = [];
  const semaphore = maxConcurrent;
  let running = 0;
  let idx = 0;
  
  const promises = dags.map(async (dag) => {
    // Wait for slot
    while (running >= semaphore) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    running++;
    const currentIdx = idx++;
    
    try {
      const success = await submitDAG(dag, apiUrl, apiKey);
      results.push({ idx: currentIdx, success });
    } catch (error) {
      results.push({ idx: currentIdx, success: false });
    } finally {
      running--;
    }
  });
  
  await Promise.all(promises);
  return results;
}

// Usage
const fileList = Array.from({ length: 1000 }, (_, i) => `uploads/file_${i}.csv`);
const dags = fileList.map((file) =>
  createContactImportDAG("my-bucket", file)[0]
);

// Submit 1000 DAGs in parallel (10 concurrent)
const results = await submitDAGsParallel(dags, 10);
const successful = results.filter(r => r.success).length;
console.log(`Submitted ${successful}/${dags.length} DAGs successfully`);
```

### Strategy 3: Sequential Submission with Rate Limit Handling

Submit DAGs sequentially, respecting rate limits.

**Python Example:**
```python
import requests
import time
from dag_builders.contact_import import create_contact_import_dag

def submit_dags_with_rate_limit(dags: list, requests_per_minute: int = 1000):
    """Submit DAGs sequentially, respecting rate limits"""
    
    api_url = "http://localhost:8000/jobs/bulk-insert/complete-graph"
    api_key = "your-api-key"
    
    # Calculate delay between requests
    delay = 60.0 / requests_per_minute  # seconds between requests
    
    results = []
    for idx, dag in enumerate(dags):
        try:
            response = requests.post(
                api_url,
                headers={"X-API-Key": api_key},
                json=dag
            )
            
            if response.status_code == 429:  # Rate limit exceeded
                print(f"Rate limit exceeded at DAG {idx}, waiting...")
                time.sleep(60)  # Wait 1 minute
                response = requests.post(
                    api_url,
                    headers={"X-API-Key": api_key},
                    json=dag
                )
            
            success = response.status_code == 200
            results.append({"idx": idx, "success": success})
            
            # Rate limiting: wait between requests
            if idx < len(dags) - 1:  # Don't wait after last request
                time.sleep(delay)
                
        except Exception as e:
            results.append({"idx": idx, "success": False, "error": str(e)})
    
    return results

# Usage
file_list = [f"uploads/file_{i}.csv" for i in range(10000)]
dags = [
    create_contact_import_dag("my-bucket", file)[0]
    for file in file_list
]

# Submit 10,000 DAGs at 1,000 requests/minute
# Estimated time: 10 minutes
results = submit_dags_with_rate_limit(dags, requests_per_minute=1000)
```

**TypeScript Example:**
```typescript
async function submitDAGsWithRateLimit(
  dags: any[][],
  requestsPerMinute: number = 1000
): Promise<Array<{ idx: number; success: boolean }>> {
  const apiUrl = "http://localhost:8000/jobs/bulk-insert/complete-graph";
  const apiKey = "your-api-key";
  
  const delay = (60 * 1000) / requestsPerMinute; // milliseconds between requests
  const results: Array<{ idx: number; success: boolean }> = [];
  
  for (let idx = 0; idx < dags.length; idx++) {
    try {
      let response = await fetch(apiUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-API-Key": apiKey
        },
        body: JSON.stringify(dags[idx])
      });
      
      if (response.status === 429) {
        // Rate limit exceeded, wait 1 minute
        console.log(`Rate limit exceeded at DAG ${idx}, waiting...`);
        await new Promise(resolve => setTimeout(resolve, 60 * 1000));
        response = await fetch(apiUrl, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-API-Key": apiKey
          },
          body: JSON.stringify(dags[idx])
        });
      }
      
      const success = response.ok;
      results.push({ idx, success });
      
      // Rate limiting: wait between requests
      if (idx < dags.length - 1) {
        await new Promise(resolve => setTimeout(resolve, delay));
      }
      
    } catch (error) {
      results.push({ idx, success: false });
    }
  }
  
  return results;
}

// Usage
const fileList = Array.from({ length: 10000 }, (_, i) => `uploads/file_${i}.csv`);
const dags = fileList.map((file) =>
  createContactImportDAG("my-bucket", file)[0]
);

// Submit 10,000 DAGs at 1,000 requests/minute
const results = await submitDAGsWithRateLimit(dags, 1000);
```

---

## Batching and Chunking Patterns

### Pattern 1: Chunk by Node Count

Split large job lists into chunks of 5,000 nodes.

**Python:**
```python
def chunk_jobs_by_count(jobs: list, chunk_size: int = 5000) -> list:
    """Split jobs into chunks of chunk_size"""
    chunks = []
    for i in range(0, len(jobs), chunk_size):
        chunks.append(jobs[i:i+chunk_size])
    return chunks

# Usage
all_jobs = [create_job(i) for i in range(50000)]
job_chunks = chunk_jobs_by_count(all_jobs, chunk_size=5000)
# Result: 10 chunks, each with 5,000 jobs

for chunk in job_chunks:
    submit_dag(chunk)
```

**TypeScript:**
```typescript
function chunkJobsByCount<T>(jobs: T[], chunkSize: number = 5000): T[][] {
  const chunks: T[][] = [];
  for (let i = 0; i < jobs.length; i += chunkSize) {
    chunks.push(jobs.slice(i, i + chunkSize));
  }
  return chunks;
}

// Usage
const allJobs = Array.from({ length: 50000 }, (_, i) => createJob(i));
const jobChunks = chunkJobsByCount(allJobs, 5000);
// Result: 10 chunks, each with 5,000 jobs

for (const chunk of jobChunks) {
  await submitDAG(chunk);
}
```

### Pattern 2: Chunk by Time Windows

Create DAGs in time-based batches (e.g., hourly, daily).

**Python:**
```python
from datetime import datetime, timedelta
from dag_builders.contact_import import create_contact_import_dag

def create_time_based_dags(file_list: list, window_size_hours: int = 1):
    """Create DAGs grouped by time windows"""
    
    # Group files by time window
    windows = {}
    for file in file_list:
        # Parse time from filename or use current time
        file_time = datetime.now()  # Or parse from filename
        window_key = file_time.replace(
            minute=0, second=0, microsecond=0
        ).isoformat()
        
        if window_key not in windows:
            windows[window_key] = []
        windows[window_key].append(file)
    
    # Create DAGs for each window
    all_dags = []
    for window_key, files in windows.items():
        window_dags = []
        for file in files:
            dag = create_contact_import_dag(
                "my-bucket",
                file,
                custom_vars={"time_window": window_key}
            )
            window_dags.extend(dag)
        
        # Split into chunks if needed
        if len(window_dags) > 5000:
            chunks = chunk_jobs_by_count(window_dags, 5000)
            all_dags.extend(chunks)
        else:
            all_dags.append(window_dags)
    
    return all_dags
```

### Pattern 3: Chunk by Dependency Groups

Split DAGs by dependency groups (connected components).

**Python:**
```python
def split_dag_by_components(dag: list) -> list:
    """Split DAG into connected components"""
    
    # Build graph
    graph = {}
    for job in dag:
        graph[job["uuid"]] = set(job.get("edges", []))
    
    # Find connected components using BFS
    visited = set()
    components = []
    
    for job in dag:
        if job["uuid"] in visited:
            continue
        
        # BFS to find connected component
        component = []
        queue = [job["uuid"]]
        visited.add(job["uuid"])
        
        while queue:
            current = queue.pop(0)
            # Find job with this UUID
            current_job = next((j for j in dag if j["uuid"] == current), None)
            if current_job:
                component.append(current_job)
            
            # Add neighbors
            for neighbor in graph.get(current, []):
                if neighbor not in visited:
                    visited.add(neighbor)
                    queue.append(neighbor)
        
        components.append(component)
    
    return components

# Usage: Split large DAG into smaller independent DAGs
large_dag = create_large_dag(10000)  # 10,000 nodes
components = split_dag_by_components(large_dag)

for component in components:
    if len(component) <= 5000:
        submit_dag(component)
    else:
        # Further split if component is too large
        chunks = chunk_jobs_by_count(component, 5000)
        for chunk in chunks:
            submit_dag(chunk)
```

---

## Scaling Strategies

### Horizontal Scaling

**API Server Scaling:**
- Multiple API server replicas share load
- Load balancer distributes requests
- Each replica has its own rate limiter

**Consumer Scaling:**
- Multiple consumers in same consumer group
- Kafka distributes partitions among consumers
- Each consumer processes jobs independently

### Vertical Scaling

**Increase Rate Limits:**
```env
# Development
MAX_REQUESTS_PER_MINUTE=1000

# Production
MAX_REQUESTS_PER_MINUTE=5000
```

**Increase Database Connections:**
```go
// clients/pgsql.go
sqldb.SetMaxOpenConns(40)  // Increase to 80, 100, etc.
sqldb.SetMaxIdleConns(20)  // Increase proportionally
```

**Add Kafka Partitions:**
- More partitions = more parallelism
- Each partition processed by one consumer
- Rebalance when consumers join/leave

### Load Distribution

**Strategy 1: Round-Robin**
```
Request 1 → API Server 1
Request 2 → API Server 2
Request 3 → API Server 3
Request 4 → API Server 1 (round-robin)
```

**Strategy 2: Least Connections**
```
Request → API Server with fewest active connections
```

**Strategy 3: Geographic Distribution**
```
Region A → API Server A (lower latency)
Region B → API Server B (lower latency)
```

---

## Best Practices

### 1. Validate Before Submission

Always validate DAG size before submission:

**Python:**
```python
from dag_builders.base import BaseDAGBuilder

dag = create_my_dag(...)

# Validate node count
if len(dag) > BaseDAGBuilder.MAX_NODES_PER_REQUEST:
    raise ValueError(f"DAG too large: {len(dag)} nodes (max: {BaseDAGBuilder.MAX_NODES_PER_REQUEST})")

# Full validation
is_valid, error = BaseDAGBuilder.validate_dag(dag)
if not is_valid:
    raise ValueError(f"DAG validation failed: {error}")
```

**TypeScript:**
```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

const dag = createMyDAG(...);

// Validate node count
if (dag.length > BaseDAGBuilder.MAX_NODES_PER_REQUEST) {
  throw new Error(`DAG too large: ${dag.length} nodes (max: ${BaseDAGBuilder.MAX_NODES_PER_REQUEST})`);
}

// Full validation
const validation = BaseDAGBuilder.validateDAG(dag);
if (!validation.isValid) {
  throw new Error(`DAG validation failed: ${validation.error}`);
}
```

### 2. Handle Rate Limits Gracefully

Implement retry logic with exponential backoff:

**Python:**
```python
import requests
import time

def submit_with_retry(dag, api_url, api_key, max_retries=3):
    """Submit DAG with retry on rate limit"""
    
    for attempt in range(max_retries):
        try:
            response = requests.post(
                api_url,
                headers={"X-API-Key": api_key},
                json=dag
            )
            
            if response.status_code == 429:
                # Rate limit exceeded
                wait_time = 2 ** attempt  # Exponential backoff: 1s, 2s, 4s
                print(f"Rate limit exceeded, waiting {wait_time}s...")
                time.sleep(wait_time)
                continue
            
            if response.status_code == 200:
                return True
            
            # Other errors
            response.raise_for_status()
            
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            time.sleep(2 ** attempt)
    
    return False
```

**TypeScript:**
```typescript
async function submitWithRetry(
  dag: any[],
  apiUrl: string,
  apiKey: string,
  maxRetries: number = 3
): Promise<boolean> {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await fetch(apiUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-API-Key": apiKey
        },
        body: JSON.stringify(dag)
      });
      
      if (response.status === 429) {
        // Rate limit exceeded
        const waitTime = Math.pow(2, attempt) * 1000; // Exponential backoff
        console.log(`Rate limit exceeded, waiting ${waitTime}ms...`);
        await new Promise(resolve => setTimeout(resolve, waitTime));
        continue;
      }
      
      if (response.ok) {
        return true;
      }
      
      throw new Error(`HTTP ${response.status}: ${await response.text()}`);
      
    } catch (error) {
      if (attempt === maxRetries - 1) {
        throw error;
      }
      await new Promise(resolve => setTimeout(resolve, Math.pow(2, attempt) * 1000));
    }
  }
  
  return false;
}
```

### 3. Monitor Submission Progress

Track DAG submission progress:

**Python:**
```python
def submit_dags_with_progress(dags: list):
    """Submit DAGs with progress tracking"""
    
    total = len(dags)
    successful = 0
    failed = 0
    
    for idx, dag in enumerate(dags):
        try:
            if submit_dag(dag):
                successful += 1
            else:
                failed += 1
        except Exception as e:
            failed += 1
            print(f"Error submitting DAG {idx}: {e}")
        
        # Progress update
        progress = ((idx + 1) / total) * 100
        print(f"Progress: {progress:.1f}% ({idx + 1}/{total}) - Success: {successful}, Failed: {failed}")
    
    return {"successful": successful, "failed": failed, "total": total}
```

### 4. Use Batch Identifiers

Tag related DAGs with batch identifiers for tracking:

**Python:**
```python
from dag_builders.contact_import import create_contact_import_dag

def create_batched_dags(file_list: list, batch_id: str):
    """Create DAGs with batch identifier"""
    
    dags = []
    for file in file_list:
        dag = create_contact_import_dag(
            "my-bucket",
            file,
            custom_vars={
                "batch_id": batch_id,
                "batch_size": len(file_list),
                "file_index": file_list.index(file)
            }
        )
        dags.extend(dag)
    
    return dags

# Usage
batch_id = f"batch-{datetime.now().strftime('%Y%m%d-%H%M%S')}"
dags = create_batched_dags(file_list, batch_id)
```

---

## Complete Examples

### Example 1: Handling 100,000 Files

**Problem**: Need to create DAGs for 100,000 files.

**Solution**: Chunk into 20 DAGs of 5,000 jobs each.

**Python:**
```python
from dag_builders.contact_import import create_contact_import_dag
import requests

def handle_large_file_batch(file_list: list):
    """Handle 100,000+ files by chunking into DAGs"""
    
    CHUNK_SIZE = 5000
    api_url = "http://localhost:8000/jobs/bulk-insert/complete-graph"
    api_key = "your-api-key"
    
    # Split files into chunks
    chunks = [file_list[i:i+CHUNK_SIZE] for i in range(0, len(file_list), CHUNK_SIZE)]
    
    print(f"Processing {len(file_list)} files in {len(chunks)} DAGs")
    
    for chunk_idx, chunk in enumerate(chunks):
        # Create DAG for this chunk
        dag = []
        for file in chunk:
            dag.extend(create_contact_import_dag(
                "my-bucket",
                file,
                custom_vars={
                    "batch_id": f"batch-{chunk_idx + 1}",
                    "chunk_index": chunk_idx,
                    "total_chunks": len(chunks)
                }
            ))
        
        # Submit DAG
        response = requests.post(
            api_url,
            headers={"X-API-Key": api_key},
            json=dag
        )
        
        if response.status_code == 200:
            print(f"✅ Submitted DAG {chunk_idx + 1}/{len(chunks)}: {len(dag)} jobs")
        else:
            print(f"❌ Failed DAG {chunk_idx + 1}/{len(chunks)}: {response.text}")
        
        # Rate limiting: wait between submissions
        if chunk_idx < len(chunks) - 1:
            time.sleep(0.1)  # 100ms delay = ~600 req/min

# Usage
file_list = [f"uploads/file_{i}.csv" for i in range(100000)]
handle_large_file_batch(file_list)
```

**TypeScript:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

async function handleLargeFileBatch(fileList: string[]) {
  const CHUNK_SIZE = 5000;
  const apiUrl = "http://localhost:8000/jobs/bulk-insert/complete-graph";
  const apiKey = "your-api-key";
  
  // Split files into chunks
  const chunks: string[][] = [];
  for (let i = 0; i < fileList.length; i += CHUNK_SIZE) {
    chunks.push(fileList.slice(i, i + CHUNK_SIZE));
  }
  
  console.log(`Processing ${fileList.length} files in ${chunks.length} DAGs`);
  
  for (let chunkIdx = 0; chunkIdx < chunks.length; chunkIdx++) {
    const chunk = chunks[chunkIdx];
    
    // Create DAG for this chunk
    const dag: any[] = [];
    chunk.forEach((file) => {
      dag.push(...createContactImportDAG("my-bucket", file, {
        customVars: {
          batch_id: `batch-${chunkIdx + 1}`,
          chunk_index: chunkIdx,
          total_chunks: chunks.length
        }
      }));
    });
    
    // Submit DAG
    const response = await fetch(apiUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-API-Key": apiKey
      },
      body: JSON.stringify(dag)
    });
    
    if (response.ok) {
      console.log(`✅ Submitted DAG ${chunkIdx + 1}/${chunks.length}: ${dag.length} jobs`);
    } else {
      console.log(`❌ Failed DAG ${chunkIdx + 1}/${chunks.length}: ${await response.text()}`);
    }
    
    // Rate limiting: wait between submissions
    if (chunkIdx < chunks.length - 1) {
      await new Promise(resolve => setTimeout(resolve, 100)); // 100ms delay
    }
  }
}

// Usage
const fileList = Array.from({ length: 100000 }, (_, i) => `uploads/file_${i}.csv`);
await handleLargeFileBatch(fileList);
```

### Example 2: Real-Time DAG Creation

**Problem**: Create DAGs as files arrive in real-time.

**Solution**: Buffer files and create DAGs in batches.

**Python:**
```python
from collections import deque
import threading
import time
from dag_builders.contact_import import create_contact_import_dag

class DAGCreator:
    def __init__(self, batch_size=5000, flush_interval=60):
        self.batch_size = batch_size
        self.flush_interval = flush_interval  # seconds
        self.buffer = deque()
        self.lock = threading.Lock()
        self.last_flush = time.time()
        
        # Start background flush thread
        self.flush_thread = threading.Thread(target=self._auto_flush, daemon=True)
        self.flush_thread.start()
    
    def add_file(self, file_path: str):
        """Add file to buffer"""
        with self.lock:
            self.buffer.append(file_path)
            
            # Flush if buffer is full
            if len(self.buffer) >= self.batch_size:
                self._flush_buffer()
    
    def _flush_buffer(self):
        """Flush buffer and create DAG"""
        if len(self.buffer) == 0:
            return
        
        with self.lock:
            files = [self.buffer.popleft() for _ in range(min(self.batch_size, len(self.buffer)))]
        
        # Create DAG
        dag = []
        for file in files:
            dag.extend(create_contact_import_dag("my-bucket", file))
        
        # Submit DAG
        submit_dag(dag)
        self.last_flush = time.time()
    
    def _auto_flush(self):
        """Auto-flush buffer at intervals"""
        while True:
            time.sleep(1)
            if time.time() - self.last_flush >= self.flush_interval:
                with self.lock:
                    if len(self.buffer) > 0:
                        self._flush_buffer()
    
    def flush(self):
        """Manually flush buffer"""
        self._flush_buffer()

# Usage
creator = DAGCreator(batch_size=5000, flush_interval=60)

# Add files as they arrive
for file in incoming_files:
    creator.add_file(file)
    
    # Or flush manually
    creator.flush()
```

---

## Task Breakdown

### Phase 1: Understanding Limits ✅

- [x] Understand MaxNodesPerRequest (5,000 nodes per DAG)
- [x] Learn API rate limits (1,000-5,000 req/min)
- [x] Study query pagination limits (100 per page)
- [x] Understand scheduler batch size (100 jobs per tick)

### Phase 2: System Capacity Analysis ✅

- [x] Calculate theoretical maximums (per minute, hour, day)
- [x] Understand practical constraints (database, Kafka, resources)
- [x] Study scalability factors (horizontal vs vertical)

### Phase 3: Large-Scale Strategies ✅

- [x] Learn chunking strategy for large DAGs
- [x] Study parallel submission patterns
- [x] Understand sequential submission with rate limiting
- [x] Learn dependency group splitting

### Phase 4: Batching Patterns ✅

- [x] Chunk by node count (5,000 nodes per chunk)
- [x] Chunk by time windows (hourly, daily)
- [x] Chunk by dependency groups (connected components)

### Phase 5: Scaling Strategies ✅

- [x] Horizontal scaling (multiple replicas)
- [x] Vertical scaling (increase limits)
- [x] Load distribution strategies

### Phase 6: Best Practices ✅

- [x] Validate before submission
- [x] Handle rate limits gracefully
- [x] Monitor submission progress
- [x] Use batch identifiers for tracking

### Phase 7: Complete Examples ✅

- [x] Study large file batch handling (100,000+ files)
- [x] Understand real-time DAG creation patterns
- [x] Practice chunking and batching strategies

---

## Summary

### How Many DAGs Can We Create?

**Short Answer**: **Unlimited** (limited by rate limits and system resources)

**Detailed Answer**:

1. **Single DAG**: Maximum **5,000 nodes** per request
2. **Multiple DAGs**: **1,000-5,000 DAGs per minute** (rate limit)
3. **Total Capacity**: **Unlimited** (no explicit limit on total number of DAGs)

### Key Strategies

1. **Chunking**: Split large DAGs (>5,000 nodes) into multiple DAGs
2. **Batching**: Group related DAGs and submit in batches
3. **Parallel Submission**: Submit multiple DAGs concurrently (respect rate limits)
4. **Sequential Submission**: Submit DAGs sequentially with rate limit handling
5. **Monitoring**: Track submission progress and handle errors gracefully

### Best Practices

1. ✅ Always validate DAG size before submission
2. ✅ Implement retry logic for rate limit errors
3. ✅ Use batch identifiers for tracking
4. ✅ Monitor submission progress
5. ✅ Scale horizontally for higher throughput

---

## Additional Resources

- [HOW_TO_CREATE_DAGS.md](./HOW_TO_CREATE_DAGS.md) - Basic DAG creation guide
- [DYNAMIC_DAGS_WITH_CUSTOM_VARS.md](./DYNAMIC_DAGS_WITH_CUSTOM_VARS.md) - Dynamic DAG creation
- [LOAD_TESTING.md](./LOAD_TESTING.md) - Performance testing guide
- [API.md](./API.md) - Complete API reference
