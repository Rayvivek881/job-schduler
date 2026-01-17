package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"vivek-ray/clients"
	"vivek-ray/conf"
	"vivek-ray/connections"
	"vivek-ray/constants"
	"vivek-ray/models"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type BaseConsumer struct {
	jobsRepo    models.JobsSvcRepo
	edgesRepo   models.EdgesSvcRepo
	kafkaReader *clients.KafkaReader
}

func NewBaseConsumer() BaseConsumerRepo {
	reader, err := connections.KafkaService.InitReader(
		conf.KafkaConfig.JobsTopic,
		conf.KafkaConfig.ConsumerGroup,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize kafka reader")
		return nil
	}

	return &BaseConsumer{
		jobsRepo:    models.JobsRepository(connections.PgDBConnection.Client),
		edgesRepo:   models.EdgesRepository(connections.PgDBConnection.Client),
		kafkaReader: reader,
	}
}

type BaseConsumerRepo interface {
	ConsumeJobs(ctx context.Context) error
}

func (b *BaseConsumer) ConsumeJobs(ctx context.Context) error {
	defer b.kafkaReader.Close()
	return b.kafkaReader.Read(ctx, b.handleMessage)
}

func (b *BaseConsumer) handleMessage(msg *sarama.ConsumerMessage) error {
	var payload struct {
		UUID string `json:"uuid"`
	}
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		log.Error().Err(err).Str("key", string(msg.Key)).Msg("failed to unmarshal message")
		return nil
	}

	job, err := b.fetchJob(payload.UUID)
	if err != nil {
		return err
	}
	job.Status = constants.ProcessingJobStatus
	switch job.JobType { // TODO: Implement job processing logic based on job.JobType
	default:
		err = errors.New("invalid job type")
	}

	if err != nil {
		job.Status = constants.FailedJobStatus
		job.RetryCount -= 1
		job.AddToJobResponse("runtime_errors", err.Error())
		job.RunAfter = time.Now().Add(time.Duration(job.RetryInterval) * time.Minute)
	} else {
		job.Status = constants.CompletedJobStatus
		err = b.edgesRepo.UpdateNodeDegree(job.UUID)
		if err != nil {
			job.Status = constants.CompletedJobStatus
			job.AddToJobResponse("runtime_errors", err.Error())
		}
	}

	return b.jobsRepo.JobsBulkUpsert([]*models.ModelJobs{job})
}

func (b *BaseConsumer) fetchJob(uuid string) (*models.ModelJobs, error) {
	jobs, err := b.jobsRepo.GetJobs(&models.JobFilters{Uuids: []string{uuid}})
	if err != nil {
		return nil, constants.ErrorWrap(constants.ErrJobFetch, err)
	}
	if len(jobs) == 0 {
		return nil, constants.ErrJobNotFound
	}

	job := jobs[0]
	if job.Status != constants.InQueueJobStatus || job.RetryCount == 0 {
		return nil, constants.ErrJobNotInQueueOrRetryCountZero
	}
	return job, nil
}
