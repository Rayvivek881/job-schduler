package processors

import (
	"context"
	"vivek-ray/models"
)

// JobProcessor defines the interface for job processors
type JobProcessor interface {
	// Process executes the job and returns an error if processing fails
	Process(ctx context.Context, job *models.ModelJobs) error
	
	// GetJobType returns the job type this processor handles
	GetJobType() string
}
