package processers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"vivek-ray/conf"
	"vivek-ray/connections"
	"vivek-ray/constants"
	"vivek-ray/models"
	companyService "vivek-ray/modules/companies/service"
	contactService "vivek-ray/modules/contacts/service"
	"vivek-ray/utilities"
)

func ExportContactsCsvToStream(writer *io.PipeWriter, vql utilities.VQLQuery) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	vql.CompanyConfig = nil
	service := contactService.NewContactService([]*models.ModelFilter{})

	if err := csvWriter.Write(vql.SelectColumns); err != nil {
		return err
	}
	remaining, unlimited := vql.Limit, vql.Limit == 0

	for unlimited || remaining > 0 {
		vql.Limit = conf.JobConfig.BatchSize
		if !unlimited {
			vql.Limit = min(remaining, conf.JobConfig.BatchSize)
		}

		contacts, err := service.ListByFilters(vql)
		if err != nil {
			return err
		}
		if len(contacts) == 0 {
			break
		}

		for _, contact := range contacts {
			row := utilities.StructToCsvSlice(contact.PgContact, vql.SelectColumns)
			if err := csvWriter.Write(row); err != nil {
				return err
			}
		}

		remaining -= len(contacts)
		vql.Cursor = contacts[len(contacts)-1].Cursor
		csvWriter.Flush()
	}
	return nil
}

func ExportCompaniesCsvToStream(writer *io.PipeWriter, vql utilities.VQLQuery) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()
	if err := csvWriter.Write(vql.SelectColumns); err != nil {
		return err
	}
	remaining, unlimited := vql.Limit, vql.Limit == 0
	service := companyService.NewCompanyService([]*models.ModelFilter{})
	for unlimited || remaining > 0 {
		vql.Limit = conf.JobConfig.BatchSize
		if !unlimited {
			vql.Limit = min(remaining, conf.JobConfig.BatchSize)
		}

		companies, err := service.ListByFilters(vql)
		if err != nil {
			return err
		}
		if len(companies) == 0 {
			break
		}

		for _, company := range companies {
			row := utilities.StructToCsvSlice(company, vql.SelectColumns)
			if err := csvWriter.Write(row); err != nil {
				return err
			}
		}

		remaining -= len(companies)
		vql.Cursor = companies[len(companies)-1].Cursor
		csvWriter.Flush()
	}
	return nil
}

func ExportCsvToStream(writer *io.PipeWriter, jobData utilities.ExportFileJobData) error {
	vql := jobData.VQL
	vql.OrderBy = []utilities.FilterOrder{{OrderBy: "uuid", OrderDirection: "desc"}}
	vql.Page = 0

	if len(vql.SelectColumns) == 0 {
		return constants.SelectColumnsRequiredError
	}

	switch jobData.Service {
	case constants.ContactsService:
		return ExportContactsCsvToStream(writer, vql)
	case constants.CompaniesService:
		return ExportCompaniesCsvToStream(writer, vql)
	default:
		return constants.InvalidServiceError
	}
}

func ProcessExportCsvFile(job *models.ModelJobNodes) error {
	var jobData utilities.ExportFileJobData
	if err := json.Unmarshal(job.Data, &jobData); err != nil {
		return err
	}
	if jobData.FileS3Bucket == "" {
		jobData.FileS3Bucket = conf.S3StorageConfig.S3Bucket
	}
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		if err := ExportCsvToStream(writer, jobData); err != nil {
			job.AddToJobResponse("runtime_errors", err.Error())
		}
	}()

	s3Key := fmt.Sprintf("%s/%s.csv", conf.S3StorageConfig.S3UploadFilePath, job.UUID)

	if err := connections.S3Connection.WriteFileStream(context.Background(), jobData.FileS3Bucket, s3Key, reader); err != nil {
		return err
	}

	job.AddToJobResponse("s3_key", s3Key)
	return nil
}
