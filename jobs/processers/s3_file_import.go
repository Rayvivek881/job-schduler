package processers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"vivek-ray/conf"
	"vivek-ray/connections"
	"vivek-ray/models"
	commonService "vivek-ray/modules/common/service"
	"vivek-ray/utilities"
)

func InsertCsvToDb(fileStream *io.ReadCloser) error {
	csvReader, batchUpsertService := csv.NewReader(*fileStream), commonService.NewBatchUpsertService()
	headers, err := csvReader.Read()
	if err != nil {
		return err
	}
	batchSize := conf.JobConfig.BatchSize
	batch := make([]map[string]string, 0, batchSize)

	for {
		row, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		batch = append(batch, utilities.CsvRowToMap(headers, row))
		if len(batch) >= batchSize {
			if err := batchUpsertService.ProcessBatchUpsert(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		return batchUpsertService.ProcessBatchUpsert(batch)
	}
	return nil
}

func ProcessInsertCsvFile(job *models.ModelJobNodes) error {
	var jobData utilities.InsertFileJobData
	if err := json.Unmarshal(job.Data, &jobData); err != nil {
		return err
	}
	if jobData.FileS3Bucket == "" {
		jobData.FileS3Bucket = conf.S3StorageConfig.S3Bucket
	}
	fileStream, err := connections.S3Connection.ReadFileStream(
		context.Background(),
		jobData.FileS3Bucket,
		jobData.FileS3Key,
	)
	if err != nil {
		return err
	}
	defer fileStream.Close()
	return InsertCsvToDb(&fileStream)
}
