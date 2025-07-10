package jobqfx

import (
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	asynqjobq "github.com/ipfs-force-community/threadmirror/pkg/jobq/asynq"
	"go.uber.org/fx"
)

var ModuleClient = fx.Module("jobqClient",
	fx.Provide(func(redisClient *redis.Client) *asynqjobq.AsynqClient {
		return asynqjobq.NewAsynqClient(redisClient)
	}),
	fx.Provide(func(c *asynqjobq.AsynqClient) jobq.JobQueueClient {
		return c
	}),
)

var ModuleServer = fx.Module("jobqServer",
	fx.Provide(func(redisClient *redis.Client, logger *slog.Logger) *asynqjobq.AsynqServer {
		return asynqjobq.NewAsynqServer(redisClient, logger)
	}),
	fx.Provide(func(s *asynqjobq.AsynqServer) jobq.JobHandlerRegistry {
		return s
	}),
	fx.Invoke(func(lc fx.Lifecycle, s *asynqjobq.AsynqServer) {
		lc.Append(fx.StartStopHook(s.Start, s.Shutdown))
	}),
)
