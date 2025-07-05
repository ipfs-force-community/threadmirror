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

// JobHandlerFunc is an adapter to allow the use of ordinary functions as JobHandler.
type JobHandlerFunc func(ctx context.Context, job *Job) error

// HandleJob calls f(ctx, job)
func (f JobHandlerFunc) HandleJob(ctx context.Context, job *Job) error {
	return f(ctx, job)
}

// JobMiddleware defines a middleware for JobHandler.
type JobMiddleware func(JobHandler) JobHandler

// ChainJobMiddlewares applies a list of middlewares to a JobHandler in order.
func ChainJobMiddlewares(handler JobHandler, mws ...JobMiddleware) JobHandler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// JobQueueClient defines the interface for enqueuing generic jobs.
type JobQueueClient interface {
	// Enqueue takes a generic job.Job and options, and returns job info.
	Enqueue(ctx context.Context, job *Job) (id string, err error)
}

type JobHandlerRegistry interface {
	RegisterHandler(jobType string, handler JobHandler)
}
