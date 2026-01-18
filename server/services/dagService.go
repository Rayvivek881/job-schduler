package services

import (
	"vivek-ray/constants"
	"vivek-ray/models"
	"vivek-ray/server/helper"

	"github.com/uptrace/bun"
)

type DAGService struct {
	jobsRepo models.JobsSvcRepo
}

func NewDAGService(db *bun.DB) DAGServiceRepo {
	return &DAGService{
		jobsRepo: models.JobsRepository(db),
	}
}

type DAGServiceRepo interface {
	GetDAGStatus(jobUUIDs []string) (*helper.DAGStatusResponse, error)
	GetDAGProgress(jobUUIDs []string) (*helper.DAGProgressResponse, error)
}

func (d *DAGService) GetDAGStatus(jobUUIDs []string) (*helper.DAGStatusResponse, error) {
	if len(jobUUIDs) == 0 {
		return &helper.DAGStatusResponse{
			DAGStatus:      "empty",
			TotalJobs:      0,
			StatusBreakdown: make(map[string]int),
			Jobs:           make(map[string]helper.JobStatusInfo),
			Progress: helper.Progress{
				Completed:  0,
				Failed:     0,
				Processing: 0,
				Pending:    0,
				Percentage: 0,
			},
		}, nil
	}

	// Get all jobs
	jobs, err := d.jobsRepo.GetJobs(&models.JobFilters{Uuids: jobUUIDs})
	if err != nil {
		return nil, err
	}

	// Group by status
	statusCounts := make(map[string]int)
	jobDetails := make(map[string]helper.JobStatusInfo)

	for _, job := range jobs {
		statusCounts[job.Status]++
		jobDetails[job.UUID] = helper.JobStatusInfo{
			UUID:      job.UUID,
			JobTitle:  job.JobTitle,
			Status:    job.Status,
			Degree:    job.Degree,
			CreatedAt: job.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: job.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	total := len(jobs)
	completed := statusCounts[constants.CompletedJobStatus]
	failed := statusCounts[constants.FailedJobStatus]
	processing := statusCounts[constants.ProcessingJobStatus]
	inQueue := statusCounts[constants.InQueueJobStatus]
	openStatus := statusCounts[constants.OpenJobStatus]
	pending := openStatus + inQueue

	// Determine overall DAG status
	dagStatus := "partial"
	if completed == total && total > 0 {
		dagStatus = "completed"
	} else if failed > 0 {
		dagStatus = "failed"
	} else if processing > 0 || inQueue > 0 {
		dagStatus = "processing"
	} else if openStatus == total && total > 0 {
		dagStatus = "pending"
	}

	percentage := 0.0
	if total > 0 {
		percentage = (float64(completed) / float64(total)) * 100
	}

	return &helper.DAGStatusResponse{
		DAGStatus:      dagStatus,
		TotalJobs:      total,
		StatusBreakdown: statusCounts,
		Jobs:           jobDetails,
		Progress: helper.Progress{
			Completed:  completed,
			Failed:     failed,
			Processing: processing,
			Pending:    pending,
			Percentage: percentage,
		},
	}, nil
}

func (d *DAGService) GetDAGProgress(jobUUIDs []string) (*helper.DAGProgressResponse, error) {
	statusResponse, err := d.GetDAGStatus(jobUUIDs)
	if err != nil {
		return nil, err
	}

	return &helper.DAGProgressResponse{
		Progress:       statusResponse.Progress,
		StatusBreakdown: statusResponse.StatusBreakdown,
		DAGStatus:      statusResponse.DAGStatus,
	}, nil
}
