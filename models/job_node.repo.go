package models

import (
	"context"
	"vivek-ray/utilities"

	"github.com/uptrace/bun"
)

type JobNodeStruct struct {
	PgDbClient *bun.DB
}

func JobNodeRepository(db *bun.DB) JobNodeSvcRepo {
	return &JobNodeStruct{
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
	if f.MinTryCount != nil {
		query = query.Where("try_count >= ?", *f.MinTryCount)
	}
	return f.DefaultFilters.ToWhere(query.Where("deleted_at IS NULL"))
}

type JobNodeSvcRepo interface {
	JobsBulkUpsert(jobs []*ModelJobNodes) error
	GetJobs(filters *JobFilters) ([]*ModelJobNodes, error)
	GetJobsCount(filters *JobFilters) (int, error)
}

func (j *JobNodeStruct) JobsBulkUpsert(jobs []*ModelJobNodes) error {
	_, err := j.PgDbClient.NewInsert().
		Model(&jobs).
		Exec(context.Background())

	return err
}

func (j *JobNodeStruct) GetJobs(filters *JobFilters) ([]*ModelJobNodes, error) {
	var jobs []*ModelJobNodes
	query := j.PgDbClient.NewSelect().
		Model(&ModelJobNodes{})

	err := filters.ToWhere(query).Scan(context.Background(), &jobs)
	return jobs, err
}

func (j *JobNodeStruct) GetJobsCount(filters *JobFilters) (int, error) {
	return filters.ToWhere(j.PgDbClient.NewSelect().Model(&ModelJobNodes{})).Count(context.Background())
}
