package jobq

import (
	"context"
)

// Job represents a generic job with a type and payload.
type Job struct {
	Type    string
	Payload []byte
}

// NewJob creates a new generic job.
func NewJob(jobType string, payload []byte) *Job {
	return &Job{
		Type:    jobType,
		Payload: payload,
	}
}

// JobHandler defines the interface for handling a specific type of generic job.Job.
type JobHandler interface {
	HandleJob(ctx context.Context, job *Job) error
}

// JobQueueClient defines the interface for enqueuing generic jobs.
type JobQueueClient interface {
	// Enqueue takes a generic job.Job and options, and returns job info.
	Enqueue(job *Job) (id string, err error)
}

type JobHandlerRegistry interface {
	RegisterHandler(jobType string, handler JobHandler)
}
