package constants

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized                  = errors.New("unauthorized: invalid or missing authentication credentials")
	ErrRateLimitExceeded             = errors.New("rate limit exceeded: too many requests, please retry after some time")
	ErrInvalidDAG                    = errors.New("validation failed: directed acyclic graph contains cycles or invalid structure")
	ErrDuplicateJobUUIDs             = errors.New("validation failed: job UUIDs must be unique across the request")
	ErrJobFetch                      = errors.New("job operation failed: unable to retrieve jobs from database")
	ErrJobStatusUpdate               = errors.New("job operation failed: unable to update job status in database")
	ErrJobNotFound                   = errors.New("job operation failed: job not found in database")
	ErrJobStatusInvalid              = errors.New("job operation failed: job status is not valid for processing")
	ErrJobNotInQueueOrRetryCountZero = errors.New("job operation failed: job is not in queue or retry count is zero")

	ErrKafkaInvalidVersion = errors.New("kafka config failed: invalid kafka version provided")
	ErrKafkaInvalidInput   = errors.New("kafka write failed: input must be a slice")
	ErrJobKafkaWrite       = errors.New("kafka write failed: unable to publish jobs to kafka")
)

func ErrorWrap(sentinel, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", sentinel, err)
}
