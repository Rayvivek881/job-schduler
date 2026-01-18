package services

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
	"vivek-ray/constants"
	"vivek-ray/models"
	"vivek-ray/server/helper"

	"github.com/uptrace/bun"
)

type JobService struct {
	jobsRepo  models.JobsSvcRepo
	edgesRepo models.EdgesSvcRepo
}

func NewJobService(db *bun.DB) JobServiceRepo {
	return &JobService{
		jobsRepo:  models.JobsRepository(db),
		edgesRepo: models.EdgesRepository(db),
	}
}

type JobServiceRepo interface {
	BulkInsertJobs(nodes []*helper.BulkInsertGraphRequest) error
	GetJobsCount(filters models.JobFilters) (int, error)
	GetJobs(filters *models.JobFilters) ([]*models.ModelJobs, error)
	InsertJobsToDB(jobs []*models.ModelJobs, edges []*models.ModelEdges) error
	UpdateAndRetriggerJob(uuid string, data json.RawMessage, retryCount *int) error
	GetJobStatusCounts(filters models.JobFilters) (map[string]int, error)
	GetJobsByStatus(status []string, filters models.JobFilters) ([]*models.ModelJobs, error)
}

func (j *JobService) GetJobsCount(filters models.JobFilters) (int, error) {
	return j.jobsRepo.GetJobsCount(&filters)
}

func (j *JobService) GetJobs(filters *models.JobFilters) ([]*models.ModelJobs, error) {
	return j.jobsRepo.GetJobs(filters)
}

func (j *JobService) InsertJobsToDB(jobs []*models.ModelJobs, edges []*models.ModelEdges) error {
	var responseErrors error
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := j.jobsRepo.JobsBulkUpsert(jobs)
		if err != nil {
			mu.Lock()
			responseErrors = errors.Join(responseErrors, err)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		err := j.edgesRepo.CreateEdges(edges)
		if err != nil {
			mu.Lock()
			responseErrors = errors.Join(responseErrors, err)
			mu.Unlock()
		}
	}()

	wg.Wait()
	return responseErrors
}

func (j *JobService) BulkInsertJobs(jobNodes []*helper.BulkInsertGraphRequest) error {
	jobUuids, server_time := make([]string, len(jobNodes)), time.Now()
	for i, node := range jobNodes {
		jobUuids[i] = node.UUID
	}
	count, err := j.GetJobsCount(models.JobFilters{Uuids: jobUuids})
	if err != nil {
		return err
	}
	if count != 0 {
		return constants.ErrDuplicateJobUUIDs
	}
	jobs, edges := make([]*models.ModelJobs, 0), make([]*models.ModelEdges, 0)

	for _, node := range jobNodes {
		for _, edge := range node.Edges {
			edges = append(edges, &models.ModelEdges{
				Source:    node.UUID,
				Target:    edge,
				CreatedAt: &server_time,
				UpdatedAt: &server_time,
			})
		}

		jobs = append(jobs, &models.ModelJobs{
			UUID:     node.UUID,
			JobTitle: node.JobTitle,
			JobType:  node.JobType,
			Degree:   node.Degree,
			Data:     node.Data,

			RetryCount:    node.RetryCount,
			RetryInterval: node.RetryInterval,
			RunAfter:      server_time,
			CreatedAt:     server_time,
			UpdatedAt:     server_time,
		})
	}
	return j.InsertJobsToDB(jobs, edges)
}

func (j *JobService) UpdateAndRetriggerJob(uuid string, data json.RawMessage, retryCount *int) error {
	jobs, err := j.jobsRepo.GetJobs(&models.JobFilters{Uuids: []string{uuid}})
	if err != nil {
		return constants.ErrorWrap(constants.ErrJobFetch, err)
	}
	if len(jobs) == 0 {
		return constants.ErrJobNotFound
	}

	job := jobs[0]

	if data != nil {
		job.Data = data
	}
	if retryCount != nil {
		job.RetryCount = *retryCount
	}

	job.Status = constants.OpenJobStatus
	job.RunAfter = time.Now()
	job.UpdatedAt = time.Now()
	job.JobResponse = nil

	return j.jobsRepo.JobsBulkUpsert([]*models.ModelJobs{job})
}

// GetJobStatusCounts returns count of jobs grouped by status
func (j *JobService) GetJobStatusCounts(filters models.JobFilters) (map[string]int, error) {
	// Get all jobs matching the filters (without status filter)
	allFilters := filters
	allFilters.Status = nil // Remove status filter to get all statuses
	
	jobs, err := j.jobsRepo.GetJobs(&allFilters)
	if err != nil {
		return nil, err
	}
	
	// Count by status
	statusCounts := make(map[string]int)
	for _, job := range jobs {
		statusCounts[job.Status]++
	}
	
	return statusCounts, nil
}

// GetJobsByStatus returns jobs filtered by status with pagination
func (j *JobService) GetJobsByStatus(status []string, filters models.JobFilters) ([]*models.ModelJobs, error) {
	filters.Status = status
	return j.jobsRepo.GetJobs(&filters)
}
