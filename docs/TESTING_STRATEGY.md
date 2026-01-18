# Testing Strategy for Job Scheduler

Comprehensive testing strategy document covering unit tests, integration tests, and edge case scenarios.

## Test Infrastructure

### Required Dependencies

Add to `go.mod`:
```go
require (
    github.com/stretchr/testify v1.8.4        // Assertions
    github.com/DATA-DOG/go-sqlmock v1.5.0     // Database mocking
)
```

### Test Structure

```
lambda/job_schduler/
├── server/
│   ├── helper/
│   │   ├── requests.go
│   │   └── requests_test.go
│   ├── services/
│   │   ├── dagService.go
│   │   ├── dagService_test.go
│   │   ├── jobService.go
│   │   └── jobService_test.go
│   └── controller/
│       └── (integration tests)
├── models/
│   ├── job_node.repo.go
│   ├── job_node.repo_test.go
│   ├── edge.repo.go
│   └── edge.repo_test.go
├── jobs/
│   ├── consumers/
│   │   ├── base.go
│   │   └── base_test.go
│   └── schduler_test.go
└── test/
    ├── fixtures/
    │   └── jobs.go
    └── mocks/
        └── kafka.go
```

## Unit Tests

### DAG Validation Tests

**File**: `server/helper/requests_test.go`

#### Test 1: Valid Linear DAG
```go
func TestVerifyDagAndUpdateDegree_ValidLinearDAG(t *testing.T) {
    dag := []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-b"}},
        {UUID: "job-b", Edges: []string{"job-c"}},
        {UUID: "job-c", Edges: []string{}},
    }
    
    result := VerifyDagAndUpdateDegree(dag)
    assert.True(t, result)
    assert.Equal(t, 0, dag[0].Degree)
    assert.Equal(t, 1, dag[1].Degree)
    assert.Equal(t, 1, dag[2].Degree)
}
```

#### Test 2: Valid Parallel DAG
```go
func TestVerifyDagAndUpdateDegree_ValidParallelDAG(t *testing.T) {
    dag := []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-b", "job-c"}},
        {UUID: "job-b", Edges: []string{}},
        {UUID: "job-c", Edges: []string{}},
    }
    
    result := VerifyDagAndUpdateDegree(dag)
    assert.True(t, result)
    assert.Equal(t, 0, dag[0].Degree)
    assert.Equal(t, 1, dag[1].Degree)
    assert.Equal(t, 1, dag[2].Degree)
}
```

#### Test 3: Cycle Detection
```go
func TestVerifyDagAndUpdateDegree_CycleDetected(t *testing.T) {
    dag := []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-b"}},
        {UUID: "job-b", Edges: []string{"job-c"}},
        {UUID: "job-c", Edges: []string{"job-a"}},  // Cycle!
    }
    
    result := VerifyDagAndUpdateDegree(dag)
    assert.False(t, result)
}
```

#### Test 4: Self-Loop Detection
```go
func TestVerifyDagAndUpdateDegree_SelfLoop(t *testing.T) {
    dag := []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-a"}},  // Self-loop
    }
    
    result := VerifyDagAndUpdateDegree(dag)
    assert.False(t, result)
}
```

#### Test 5: Missing Target Node
```go
func TestVerifyDagAndUpdateDegree_MissingTarget(t *testing.T) {
    dag := []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-b"}},  // job-b doesn't exist
    }
    
    result := VerifyDagAndUpdateDegree(dag)
    assert.False(t, result)
}
```

#### Test 6: Large DAG (1000 nodes)
```go
func TestVerifyDagAndUpdateDegree_LargeDAG(t *testing.T) {
    dag := make([]*BulkInsertGraphRequest, 1000)
    for i := 0; i < 1000; i++ {
        dag[i] = &BulkInsertGraphRequest{
            UUID: fmt.Sprintf("job-%d", i),
            Edges: []string{},
        }
        if i > 0 {
            dag[i-1].Edges = []string{dag[i].UUID}
        }
    }
    
    result := VerifyDagAndUpdateDegree(dag)
    assert.True(t, result)
    
    // Verify degrees
    assert.Equal(t, 0, dag[0].Degree)
    for i := 1; i < 1000; i++ {
        assert.Equal(t, 1, dag[i].Degree)
    }
}
```

### Repository Tests

**File**: `models/job_node.repo_test.go`

#### Test: GetJobs with Filters
```go
func TestJobsStruct_GetJobs(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()
    
    bunDB := bun.NewDB(sql.OpenDB("pgx", db), pgdialect.New())
    repo := JobsRepository(bunDB)
    
    filters := &JobFilters{
        Status: []string{"open"},
        Degree: func() *int { d := 0; return &d }(),
    }
    
    rows := sqlmock.NewRows([]string{"id", "uuid", "status", "degree"}).
        AddRow(1, "job-a", "open", 0)
    
    mock.ExpectQuery("SELECT (.+) FROM jobs").
        WithArgs("open", 0).
        WillReturnRows(rows)
    
    jobs, err := repo.GetJobs(filters)
    assert.NoError(t, err)
    assert.Len(t, jobs, 1)
    assert.Equal(t, "job-a", jobs[0].UUID)
}
```

### Consumer Tests

**File**: `jobs/consumers/base_test.go`

#### Test: Successful Job Processing
```go
func TestBaseConsumer_handleMessage_Success(t *testing.T) {
    // Mock job repository
    // Mock processor (return success)
    // Verify status update to 'completed'
    // Verify degree decrement called
}
```

#### Test: Failed Job Processing
```go
func TestBaseConsumer_handleMessage_Failure(t *testing.T) {
    // Mock job repository
    // Mock processor (return error)
    // Verify status update to 'failed'
    // Verify retry_count decrement
    // Verify run_after update
}
```

#### Test: Timeout Handling
```go
func TestBaseConsumer_handleMessage_Timeout(t *testing.T) {
    // Mock processor with timeout
    // Verify context.DeadlineExceeded handling
    // Verify timeout error in job_response
}
```

## Integration Tests

### End-to-End DAG Execution

**File**: `test/integration/dag_execution_test.go`

```go
func TestIntegration_CompleteDAGExecution(t *testing.T) {
    // 1. Create DAG via API
    // 2. Verify jobs inserted with correct degrees
    // 3. Start scheduler (mock or test instance)
    // 4. Verify jobs published to Kafka
    // 5. Process jobs via consumer
    // 6. Verify dependency cascade
    // 7. Verify all jobs completed
}
```

### Retry Mechanism Test

```go
func TestIntegration_JobRetry(t *testing.T) {
    // 1. Create job with retry_count=2
    // 2. Mock processor to fail first time
    // 3. Verify job marked as failed
    // 4. Wait for retry_interval
    // 5. Verify retry scheduler picks up job
    // 6. Mock processor to succeed second time
    // 7. Verify job completed
}
```

## Edge Case Tests

### Race Condition Tests

```go
func TestEdgeCase_MultipleSchedulerRace(t *testing.T) {
    // Simulate two schedulers querying simultaneously
    // Verify only one processes each job
}
```

### Stale Job Test

```go
func TestEdgeCase_StaleProcessingJob(t *testing.T) {
    // Create job with status='processing'
    // updated_at = 2 hours ago
    // Verify stale job detection
    // Verify recovery mechanism
}
```

## Performance Tests

### Benchmark DAG Validation

```go
func BenchmarkVerifyDagAndUpdateDegree_Large(b *testing.B) {
    dag := generateLargeDAG(10000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        VerifyDagAndUpdateDegree(dag)
    }
}
```

## Test Fixtures

**File**: `test/fixtures/jobs.go`

```go
func CreateTestLinearDAG() []*BulkInsertGraphRequest {
    return []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-b"}},
        {UUID: "job-b", Edges: []string{"job-c"}},
        {UUID: "job-c", Edges: []string{}},
    }
}

func CreateTestParallelDAG() []*BulkInsertGraphRequest {
    return []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-b", "job-c"}},
        {UUID: "job-b", Edges: []string{}},
        {UUID: "job-c", Edges: []string{}},
    }
}

func CreateTestCycleDAG() []*BulkInsertGraphRequest {
    return []*BulkInsertGraphRequest{
        {UUID: "job-a", Edges: []string{"job-b"}},
        {UUID: "job-b", Edges: []string{"job-c"}},
        {UUID: "job-c", Edges: []string{"job-a"}},  // Cycle
    }
}
```

## Test Coverage Goals

- **DAG Validation**: 100% coverage
- **Repository Layer**: 80%+ coverage
- **Service Layer**: 80%+ coverage
- **Consumer Logic**: 80%+ coverage
- **Scheduler Logic**: 80%+ coverage

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestVerifyDagAndUpdateDegree_ValidLinearDAG ./server/helper
```
