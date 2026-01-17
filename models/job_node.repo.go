package models

import (
	"context"
	"vivek-ray/utilities"

	"github.com/uptrace/bun"
)

type JobsStruct struct {
	PgDbClient *bun.DB
}

func JobsRepository(db *bun.DB) JobsSvcRepo {
	return &JobsStruct{
		PgDbClient: db,
	}
}

func (f *JobFilters) ToWhere(query *bun.SelectQuery) *bun.SelectQuery {
	if len(f.Uuids) > 0 {
		query = query.Where("uuid IN (?)", bun.In(utilities.UniqueStringSlice(f.Uuids)))
	}
	if f.Degree != nil {
		query = query.Where("degree = ?", *f.Degree)
	}
	if f.RunAfter != nil {
		query = query.Where("run_after <= ?", *f.RunAfter)
	}
	if len(f.Status) > 0 {
		query = query.Where("status IN (?)", bun.In(f.Status))
	}
	return f.DefaultFilters.ToWhere(query)
}

type JobsSvcRepo interface {
	JobsBulkUpsert(jobs []*ModelJobs) error
	GetJobs(filters *JobFilters) ([]*ModelJobs, error)
	GetJobsCount(filters *JobFilters) (int, error)
}

func (j *JobsStruct) JobsBulkUpsert(jobs []*ModelJobs) error {
	_, err := j.PgDbClient.NewInsert().
		Model(&jobs).
		Exec(context.Background())

	return err
}

func (j *JobsStruct) GetJobs(filters *JobFilters) ([]*ModelJobs, error) {
	var jobs []*ModelJobs
	query := j.PgDbClient.NewSelect().
		Model(&ModelJobs{})

	err := filters.ToWhere(query).Scan(context.Background(), &jobs)
	return jobs, err
}

func (j *JobsStruct) GetJobsCount(filters *JobFilters) (int, error) {
	return filters.ToWhere(j.PgDbClient.NewSelect().Model(&ModelJobs{})).Count(context.Background())
}
