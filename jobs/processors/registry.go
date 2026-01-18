package processors

import (
	"context"
	"fmt"
	"sync"
	"vivek-ray/constants"
	"vivek-ray/models"

	"github.com/rs/zerolog/log"
)

var (
	processors      sync.Map
	once            sync.Once
	defaultProcessor JobProcessor
)

// Init initializes the processor registry
func Init() {
	once.Do(func() {
		defaultProcessor = NewDefaultProcessor()
		
		// Register default processor for generic job types
		Register(defaultProcessor)
		
		log.Info().Msg("Job processor registry initialized")
	})
}

// Register registers a job processor for a specific job type
func Register(processor JobProcessor) {
	if processor == nil {
		log.Warn().Msg("Attempted to register nil processor, skipping")
		return
	}
	
	jobType := processor.GetJobType()
	processors.Store(jobType, processor)
	log.Info().Str("job_type", jobType).Msg("Registered job processor")
}

// GetProcessor returns the processor for the given job type, or default if not found
func GetProcessor(jobType string) JobProcessor {
	if processor, exists := processors.Load(jobType); exists {
		return processor.(JobProcessor)
	}
	
	log.Warn().Str("job_type", jobType).Msg("No processor found for job type, using default")
	return defaultProcessor
}

// ProcessJob processes a job using the appropriate processor
func ProcessJob(ctx context.Context, job *models.ModelJobs) error {
	if job == nil {
		return constants.ErrJobProcessorNotFound
	}
	
	processor := GetProcessor(job.JobType)
	if processor == nil {
		return fmt.Errorf("%w: job type %s", constants.ErrJobProcessorNotFound, job.JobType)
	}
	
	return processor.Process(ctx, job)
}
