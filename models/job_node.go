package models

import (
	"encoding/json"
	"time"
	"vivek-ray/utilities"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type JobFilters struct {
	Uuids       []string   `json:"uuids"`
	Degree      *int       `json:"degree"`
	RunAfter    *time.Time `json:"run_after"`
	MinTryCount *int       `json:"min_try_count"`
	Status      []string   `json:"status"`

	utilities.DefaultFilters
}

type JobResponseStruct struct {
	Message       string   `json:"message"`
	RuntimeErrors []string `json:"runtime_errors"`
	S3Key         string   `json:"s3_key"`
}

type ModelJobNodes struct {
	db            *bun.DB
	bun.BaseModel `bun:"table:job_node,alias:jn"`

	Id          uint64          `bun:"id,pk,autoincrement" json:"id"`
	UUID        string          `bun:"uuid,notnull,unique" json:"uuid"`
	JobTitle    string          `bun:"job_title,notnull" json:"job_title"`
	JobType     string          `bun:"job_type,notnull" json:"job_type"`
	Degree      int             `bun:"degree,notnull,default:0" json:"degree"`
	Data        json.RawMessage `bun:"data,type:jsonb,default:'{}'" json:"data"`
	Status      string          `bun:"status,notnull,default:'open'" json:"status"`
	JobResponse json.RawMessage `bun:"job_response,type:jsonb,default:'{}'" json:"job_response"`

	TryCount      int        `bun:"try_count,notnull,default:1" json:"try_count"`
	RetryInterval int        `bun:"retry_interval,notnull,default:30" json:"retry_interval"`
	RunAfter      time.Time  `bun:"run_after,nullzero,default:current_timestamp" json:"run_after"`
	DeletedAt     *time.Time `bun:"deleted_at,nullzero" json:"deleted_at"`

	CreatedAt time.Time `bun:"created_at,nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,default:current_timestamp" json:"updated_at"`
}

func (m *ModelJobNodes) SetDB(db *bun.DB) *ModelJobNodes {
	m.db = db
	return m
}

func (m *ModelJobNodes) AddToJobResponse(key, value string) {
	jobResponse := JobResponseStruct{}
	if m.JobResponse != nil {
		if err := json.Unmarshal(m.JobResponse, &jobResponse); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal job response")
			return
		}
	}

	switch key {
	case "message":
		jobResponse.Message = value
	case "runtime_errors":
		if jobResponse.RuntimeErrors == nil {
			jobResponse.RuntimeErrors = make([]string, 0)
		}
		jobResponse.RuntimeErrors = append(jobResponse.RuntimeErrors, value)
	case "s3_key":
		jobResponse.S3Key = value
	}
	data, err := json.Marshal(jobResponse)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal job response")
		return
	}
	m.JobResponse = data
}
