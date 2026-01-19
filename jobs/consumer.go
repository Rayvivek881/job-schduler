package jobs

import (
	"context"
	"encoding/json"
	"time"
	"vivek-ray/clients"
	"vivek-ray/conf"
	"vivek-ray/connections"
	"vivek-ray/constants"
	"vivek-ray/jobs/processers"
	"vivek-ray/models"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type BaseConsumer struct {
	jobsRepo    models.JobNodeSvcRepo
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
		jobsRepo:    models.JobNodeRepository(connections.PgDBConnection.Client),
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
	err = b.processJob(job)

	if err != nil {
		job.Status = constants.FailedJobStatus
		job.TryCount -= 1
		job.AddToJobResponse("runtime_errors", err.Error())
		job.RunAfter = time.Now().Add(time.Duration(job.RetryInterval) * time.Minute)

	} else {
		job.Status = constants.CompletedJobStatus
		err = b.edgesRepo.UpdateNodeDegree(job.UUID)
		if err != nil {
			job.AddToJobResponse("runtime_errors", err.Error())
		}
	}

	job.UpdatedAt = time.Now()
	return b.jobsRepo.JobsBulkUpsert([]*models.ModelJobNodes{job})
}

func (b *BaseConsumer) fetchJob(uuid string) (*models.ModelJobNodes, error) {
	jobs, err := b.jobsRepo.GetJobs(&models.JobFilters{Uuids: []string{uuid}})
	if err != nil {
		return nil, constants.ErrorWrap(constants.ErrJobFetch, err)
	}
	if len(jobs) == 0 {
		return nil, constants.ErrJobNotFound
	}

	job := jobs[0]
	if job.Status != constants.InQueueJobStatus || job.TryCount == 0 {
		return nil, constants.ErrJobNotInQueueOrTryCountZero
	}
	return job, nil
}

func (b *BaseConsumer) processJob(job *models.ModelJobNodes) error {
	switch job.JobType {
	case constants.InsertCsvFile:
		return processers.ProcessInsertCsvFile(job)
	case constants.ExportCsvFile:
		return processers.ProcessExportCsvFile(job)
	default:
		return constants.InvalidJobTypeError(job.JobType)
	}
}
