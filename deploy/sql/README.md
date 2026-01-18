# Database Schema Scripts

This directory contains SQL scripts for initializing the Job Scheduler database schema.

## Files

- `init_schema.sql` - Main schema initialization script

## Usage

### Initialize Database Schema

```bash
# Using psql command line
psql -U postgres -d job_scheduler -f deploy/sql/init_schema.sql

# Or using environment variable
PGPASSWORD=your_password psql -h localhost -U postgres -d job_scheduler -f deploy/sql/init_schema.sql
```

### Verify Schema

```bash
# List tables
psql -U postgres -d job_scheduler -c "\dt"

# Describe jobs table
psql -U postgres -d job_scheduler -c "\d jobs"

# Describe edges table
psql -U postgres -d job_scheduler -c "\d edges"

# List indexes
psql -U postgres -d job_scheduler -c "\di"
```

## Tables

### jobs
Stores job definitions and execution state.

**Key Columns:**
- `uuid` - Unique job identifier (PK)
- `degree` - In-degree count (0 = ready to execute)
- `status` - Job status: `open`, `in_queue`, `processing`, `completed`, `failed`
- `data` - JSONB payload for custom job data
- `job_response` - JSONB for execution results/errors

### edges
Represents DAG dependencies between jobs.

**Key Columns:**
- `source` - Source job UUID (must complete first)
- `target` - Target job UUID (waits for source)

**Constraints:**
- `source` → `target` (source must complete before target)
- Foreign keys to `jobs.uuid` with CASCADE delete
- Unique constraint on (source, target) pairs
- No self-referential edges allowed

## Indexes

The schema includes optimized indexes for:

1. **Scheduler queries**: `idx_jobs_degree_status_runafter` - Fast lookup of ready jobs
2. **UUID lookups**: `idx_jobs_uuid` - Fast job retrieval by UUID
3. **Status filtering**: `idx_jobs_status` - Filter jobs by status
4. **Edge queries**: `idx_edges_source` and `idx_edges_target` - Fast dependency lookups

## Notes

- The schema uses soft deletes (`deleted_at` column)
- Foreign keys use CASCADE delete to maintain referential integrity
- Indexes include partial indexes with `WHERE deleted_at IS NULL` for better performance
