package constants

import "time"

var (
	OpenJobStatus       = "open"
	InQueueJobStatus    = "in_queue"
	ProcessingJobStatus = "processing"
	CompletedJobStatus  = "completed"
	FailedJobStatus     = "failed"

	RetryingJobStatus  = "retrying"
	MaxNodesPerRequest = 5000
	MaxPageLimit       = 100

	FirstTimeJobType = "first_time"
	RetryJobType     = "retry"

	// Default job execution timeout (30 minutes)
	DefaultJobExecutionTimeout = 30 * time.Minute
)
