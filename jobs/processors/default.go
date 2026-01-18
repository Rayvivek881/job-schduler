package processors

import (
	"context"
	"encoding/json"
	"vivek-ray/models"

	"github.com/rs/zerolog/log"
)

const DefaultJobType = "default"

type DefaultProcessor struct{}

func NewDefaultProcessor() *DefaultProcessor {
	return &DefaultProcessor{}
}

func (p *DefaultProcessor) GetJobType() string {
	return DefaultJobType
}

func (p *DefaultProcessor) Process(ctx context.Context, job *models.ModelJobs) error {
	log.Info().
		Str("job_uuid", job.UUID).
		Str("job_type", job.JobType).
		Str("job_title", job.JobTitle).
		Msg("Processing job with default processor")
	
	// Parse job data
	var jobData map[string]interface{}
	if len(job.Data) > 0 {
		if err := json.Unmarshal(job.Data, &jobData); err != nil {
			log.Warn().Err(err).Msg("Failed to parse job data, continuing with empty data")
		}
	}
	
	// Default implementation: just log the job
	// Users can extend this or register custom processors
	log.Info().
		Str("job_uuid", job.UUID).
		Interface("job_data", jobData).
		Msg("Job processed successfully (default processor)")
	
	// Set success message in job response
	job.AddToJobResponse("message", "Job processed successfully by default processor")
	
	return nil
}
