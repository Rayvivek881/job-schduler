package jobs

import (
	"context"
	"time"
	"vivek-ray/clients"
	"vivek-ray/conf"
	"vivek-ray/connections"
	"vivek-ray/constants"
	"vivek-ray/models"
	"vivek-ray/utilities"

	"github.com/rs/zerolog/log"
)

type JobScheduler struct {
	jobsRepo    models.JobsSvcRepo
	edgesRepo   models.EdgesSvcRepo
	kafkaWriter *clients.KafkaWriter
}

func NewJobScheduler() JobSvc {
	kafkaWriter, err := connections.KafkaService.InitWriter(conf.KafkaConfig.JobsTopic)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing kafka writer")
		return nil
	}

	return &JobScheduler{
		jobsRepo:    models.JobsRepository(connections.PgDBConnection.Client),
		edgesRepo:   models.EdgesRepository(connections.PgDBConnection.Client),
		kafkaWriter: kafkaWriter,
	}
}

type JobSvc interface {
	FirstTimeJobs(ctx context.Context, ticker *time.Ticker)
	RetryJobs(ctx context.Context, ticker *time.Ticker)
	InsertJobsToKafka(jobFilters *models.JobFilters) error
}

func (j *JobScheduler) InsertJobsToKafka(jobFilters *models.JobFilters) error {
	limit := constants.MaxPageLimit

	for {
		jobFilters.DefaultFilters = utilities.DefaultFilters{
			Limit: limit,
			Order: []*utilities.OrderByDirection{{Column: "run_after", Direction: "ASC"}},
		}

		jobs, err := j.jobsRepo.GetJobs(jobFilters)
		if err != nil {
			return constants.ErrorWrap(constants.ErrJobFetch, err)
		}
		if len(jobs) == 0 {
			break
		}
		for _, job := range jobs {
			job.Status = constants.InQueueJobStatus
		}
		if err = j.jobsRepo.JobsBulkUpsert(jobs); err != nil {
			return constants.ErrorWrap(constants.ErrJobStatusUpdate, err)
		}

		if err = j.kafkaWriter.BulkWrite(jobs); err != nil {
			return constants.ErrorWrap(constants.ErrJobKafkaWrite, err)
		}
	}
	return nil
}

func (j *JobScheduler) FirstTimeJobs(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			j.kafkaWriter.Close()
			return
		case <-ticker.C:
			curr_time := time.Now()
			jobFilter := &models.JobFilters{
				Status:   []string{constants.OpenJobStatus},
				RunAfter: &curr_time,
			}
			err := j.InsertJobsToKafka(jobFilter)
			if err != nil {
				log.Error().Err(err).Msg("Error in InJobsToKafka")
			}
		}
	}
}

func (j *JobScheduler) RetryJobs(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			j.kafkaWriter.Close()
			return
		case <-ticker.C:
			curr_time := time.Now()
			jobFilter := &models.JobFilters{
				Status:   []string{constants.FailedJobStatus},
				RunAfter: &curr_time,
			}
			err := j.InsertJobsToKafka(jobFilter)
			if err != nil {
				log.Error().Err(err).Msg("Error in InJobsToKafka")
			}
		}
	}
}

func RunJobs(ctx context.Context, args []string) {
	if len(args) == 0 {
		log.Error().Msg("Job type is required")
		return
	}
	jobType, jobScheduler := args[0], NewJobScheduler()
	ticker := time.NewTicker(time.Duration(conf.JobConfig.TickerInterval) * time.Minute)
	defer ticker.Stop()

	switch jobType {
	case constants.FirstTimeJobType:
		jobScheduler.FirstTimeJobs(ctx, ticker)
	case constants.RetryJobType:
		jobScheduler.RetryJobs(ctx, ticker)
	default:
		log.Error().Msg("Invalid job type")
	}
}
