package asynqjobq

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/redis/go-redis/v9"
)

// AsynqClient implements job.JobQueueClient for Asynq.
type AsynqClient struct {
	*asynq.Client
	defaultOptions []asynq.Option
}

// NewAsynqClient creates a new AsynqClient.
func NewAsynqClient(redisClient redis.UniversalClient, defaultOptions ...asynq.Option) *AsynqClient {
	return &AsynqClient{
		Client:         asynq.NewClientFromRedisClient(redisClient),
		defaultOptions: defaultOptions,
	}
}

// Enqueue enqueues a job to Asynq.
func (c *AsynqClient) Enqueue(ctx context.Context, job *jobq.Job) (string, error) {
	asynqTask := asynq.NewTask(job.Type, job.Payload)
	taskInfo, err := c.EnqueueContext(ctx, asynqTask, c.defaultOptions...)
	if err != nil {
		return "", err
	}
	return taskInfo.ID, nil
}

// AsynqServer implements job.JobQueueServer for Asynq.
type AsynqServer struct {
	*asynq.Server
	mux    *asynq.ServeMux
	logger *slog.Logger
}

// NewAsynqServer creates a new AsynqServer.
func NewAsynqServer(redisClient redis.UniversalClient, logger *slog.Logger) *AsynqServer {
	mux := asynq.NewServeMux()
	server := asynq.NewServerFromRedisClient(
		redisClient,
		asynq.Config{
			Concurrency: 1,
			Queues: map[string]int{
				"default": 1,
			},
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("Job processing failed",
					"job_type", task.Type(),
					"payload_size", len(task.Payload()),
					"error", err,
				)
			}),
			HealthCheckFunc: func(err error) {
				if err != nil {
					logger.Error("Asynq health check failed", "error", err)
				}
			},
			LogLevel: asynq.InfoLevel,
			BaseContext: func() context.Context {
				return context.Background()
			},
		},
	)

	return &AsynqServer{
		Server: server,
		mux:    mux,
		logger: logger,
	}
}

// Start starts the Asynq server.
func (s *AsynqServer) Start() error {
	return s.Server.Start(s.mux)
}

func (s *AsynqServer) RegisterHandler(jobType string, handler jobq.JobHandler) {
	s.mux.HandleFunc(jobType, withLogging(s.logger, func(ctx context.Context, t *asynq.Task) error {
		jobJob := &jobq.Job{
			Type:    t.Type(),
			Payload: t.Payload(),
		}
		return handler.HandleJob(ctx, jobJob)
	}))
}

// withLogging is a middleware that wraps job handlers with logging.
func withLogging(logger *slog.Logger, handler asynq.HandlerFunc) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		start := time.Now()

		jobLogger := logger.With(
			"job_type", task.Type(),
			"job_payload_size", len(task.Payload()),
		)

		jobLogger.Debug("Starting job processing")

		err := handler(ctx, task)

		duration := time.Since(start)

		if err != nil {
			jobLogger.Error("Job processing failed",
				"error", err,
				"duration", duration,
			)
		} else {
			jobLogger.Info("Job processing completed successfully",
				"duration", duration,
			)
		}

		return err
	}
}
