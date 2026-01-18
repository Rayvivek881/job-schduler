-- Job Scheduler Database Schema
-- PostgreSQL 14+ Required
-- 
-- This script creates the tables and indexes required for the Job Scheduler system.
-- Tables: jobs, edges
--
-- Execute this script to initialize the database schema:
--   psql -U postgres -d job_scheduler -f deploy/sql/init_schema.sql

-- ============================================================================
-- JOBS TABLE
-- ============================================================================
-- Stores job definitions and execution state
-- Key fields: uuid (PK), degree (in-degree for DAG), status, data (JSONB payload)

CREATE TABLE IF NOT EXISTS jobs (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(255) NOT NULL UNIQUE,
    job_title VARCHAR(255) NOT NULL,
    job_type VARCHAR(255) NOT NULL,
    degree INTEGER NOT NULL DEFAULT 0,
    data JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    job_response JSONB DEFAULT '{}'::jsonb,
    
    retry_count INTEGER NOT NULL DEFAULT 0,
    retry_interval INTEGER NOT NULL DEFAULT 30,
    run_after TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- EDGES TABLE
-- ============================================================================
-- Represents DAG dependencies: source job must complete before target job
-- Relationship: source → target (source completes → target becomes eligible)

CREATE TABLE IF NOT EXISTS edges (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    target VARCHAR(255) NOT NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints (ON DELETE CASCADE to clean up edges when jobs are deleted)
    CONSTRAINT fk_edges_source FOREIGN KEY (source) REFERENCES jobs(uuid) ON DELETE CASCADE,
    CONSTRAINT fk_edges_target FOREIGN KEY (target) REFERENCES jobs(uuid) ON DELETE CASCADE,
    
    -- Prevent self-referential edges and duplicate edges
    CONSTRAINT chk_edges_no_self_reference CHECK (source != target),
    CONSTRAINT uk_edges_source_target UNIQUE (source, target)
);

-- ============================================================================
-- INDEXES
-- ============================================================================
-- Optimize common query patterns

-- Composite index for scheduler queries: find ready jobs (degree=0, status='open', run_after <= NOW())
-- This is the most critical query for performance
CREATE INDEX IF NOT EXISTS idx_jobs_degree_status_runafter 
    ON jobs(degree, status, run_after) 
    WHERE deleted_at IS NULL;

-- Index for UUID lookups (used in job retrieval)
CREATE INDEX IF NOT EXISTS idx_jobs_uuid 
    ON jobs(uuid);

-- Index for status filtering (used in job status queries)
CREATE INDEX IF NOT EXISTS idx_jobs_status 
    ON jobs(status) 
    WHERE deleted_at IS NULL;

-- Index for retry scheduler queries (failed jobs with retry_count > 0)
CREATE INDEX IF NOT EXISTS idx_jobs_status_runafter 
    ON jobs(status, run_after) 
    WHERE status = 'failed' AND retry_count > 0 AND deleted_at IS NULL;

-- Index for edge source lookups (used when decrementing dependent jobs' degrees)
CREATE INDEX IF NOT EXISTS idx_edges_source 
    ON edges(source);

-- Index for edge target lookups (used in dependency queries)
CREATE INDEX IF NOT EXISTS idx_edges_target 
    ON edges(target);

-- Index for job_type filtering (if needed for job type queries)
CREATE INDEX IF NOT EXISTS idx_jobs_job_type 
    ON jobs(job_type) 
    WHERE deleted_at IS NULL;

-- Index for created_at (useful for time-based queries and monitoring)
CREATE INDEX IF NOT EXISTS idx_jobs_created_at 
    ON jobs(created_at DESC) 
    WHERE deleted_at IS NULL;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE jobs IS 'Stores job definitions, execution state, and metadata for the DAG job scheduler';
COMMENT ON TABLE edges IS 'Represents DAG dependencies: source job must complete before target job can execute';

COMMENT ON COLUMN jobs.uuid IS 'Unique identifier for the job';
COMMENT ON COLUMN jobs.degree IS 'In-degree: number of dependencies this job has (0 = ready to execute)';
COMMENT ON COLUMN jobs.status IS 'Job status: open, in_queue, processing, completed, failed';
COMMENT ON COLUMN jobs.data IS 'Job payload as JSONB (custom data for processors)';
COMMENT ON COLUMN jobs.job_response IS 'Job execution results/errors as JSONB';
COMMENT ON COLUMN jobs.retry_count IS 'Number of retries remaining (decremented on failure)';
COMMENT ON COLUMN jobs.retry_interval IS 'Minutes to wait before retry';
COMMENT ON COLUMN jobs.run_after IS 'Earliest time this job can be executed';

COMMENT ON COLUMN edges.source IS 'Source job UUID (must complete before target)';
COMMENT ON COLUMN edges.target IS 'Target job UUID (waits for source to complete)';
