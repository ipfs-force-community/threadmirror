package cronfx

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/task/cron"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"go.uber.org/fx"
)

// Module provides the cron dependency injection module
var Module = fx.Module("cron",
	fx.Provide(newCronScheduler),
	fx.Provide(newThreadStatusCleanupHandler),
	fx.Provide(newMentionCheckHandler),
	fx.Invoke(registerCronLifecycle),
)

// newCronScheduler creates a new cron scheduler
func newCronScheduler(logger *slog.Logger) (gocron.Scheduler, error) {
	return gocron.NewScheduler()
}

// newThreadStatusCleanupHandler creates a thread status cleanup handler
func newThreadStatusCleanupHandler(
	logger *slog.Logger,
	threadService *service.ThreadService,
	jobQueueClient jobq.JobQueueClient,
	cronConfig *config.CronConfig,
) *cron.ThreadStatusCleanupHandler {
	cleanupConfig := cron.ThreadStatusCleanupConfig{
		ScrapingTimeoutMinutes: cronConfig.ThreadStatusCleanup.ScrapingTimeoutMinutes,
		PendingTimeoutMinutes:  cronConfig.ThreadStatusCleanup.PendingTimeoutMinutes,
		RetryDelayMinutes:      cronConfig.ThreadStatusCleanup.RetryDelayMinutes,
		MaxRetries:             cronConfig.ThreadStatusCleanup.MaxRetries,
	}

	return cron.NewThreadStatusCleanupHandler(
		logger,
		threadService,
		jobQueueClient,
		cleanupConfig,
	)
}

// newMentionCheckHandler creates a mention check handler
func newMentionCheckHandler(
	logger *slog.Logger,
	scrapers []*xscraper.XScraper,
	jobQueueClient jobq.JobQueueClient,
	cronConfig *config.CronConfig,
) *cron.MentionCheckHandler {
	mentionConfig := cron.MentionCheckConfig{
		ExcludeMentionAuthorPrefix: cronConfig.MentionCheck.ExcludeMentionAuthorPrefix,
		MentionUsername:            cronConfig.MentionCheck.MentionUsername,
	}

	return cron.NewMentionCheckHandler(
		logger,
		scrapers,
		jobQueueClient,
		mentionConfig,
	)
}

// registerCronLifecycle registers cron jobs and manages their lifecycle
func registerCronLifecycle(
	lc fx.Lifecycle,
	scheduler gocron.Scheduler,
	threadStatusCleanup *cron.ThreadStatusCleanupHandler,
	mentionCheck *cron.MentionCheckHandler,
	cronConfig *config.CronConfig,
	logger *slog.Logger,
) {
	lc.Append(fx.StartStopHook(
		func(ctx context.Context) error {
			logger.Info("Starting cron scheduler")

			// Schedule thread status cleanup
			if cronConfig.ThreadStatusCleanup.EnabledIntervalMinutes > 0 {
				intervalMinutes := cronConfig.ThreadStatusCleanup.EnabledIntervalMinutes

				_, err := scheduler.NewJob(
					gocron.DurationJob(time.Duration(intervalMinutes)*time.Minute),
					gocron.NewTask(func() {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
						defer cancel()

						if err := threadStatusCleanup.Execute(ctx); err != nil {
							logger.Error("Thread status cleanup failed", "error", err)
						}
					}),
				)
				if err != nil {
					return err
				}
				logger.Info("Scheduled thread status cleanup", "interval_minutes", intervalMinutes)
			}

			// Schedule mention check with random intervals
			if cronConfig.MentionCheck.EnabledIntervalMinutes > 0 {
				baseInterval := time.Duration(cronConfig.MentionCheck.EnabledIntervalMinutes) * time.Minute
				randomizeRange := time.Duration(cronConfig.MentionCheck.RandomizeIntervalMinutes) * time.Minute

				_, err := scheduler.NewJob(
					gocron.DurationRandomJob(baseInterval, baseInterval+randomizeRange),
					gocron.NewTask(func() {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
						defer cancel()

						if err := mentionCheck.Execute(ctx); err != nil {
							logger.Error("Mention check failed", "error", err)
						}
					}),
				)
				if err != nil {
					return err
				}
				logger.Info("Scheduled mention check with random intervals",
					"base_interval_minutes", cronConfig.MentionCheck.EnabledIntervalMinutes,
					"randomize_range_minutes", cronConfig.MentionCheck.RandomizeIntervalMinutes,
				)
			}

			// Start the scheduler
			scheduler.Start()
			logger.Info("Cron scheduler started")

			return nil
		},
		func(ctx context.Context) error {
			logger.Info("Stopping cron scheduler")
			if err := scheduler.Shutdown(); err != nil {
				logger.Error("Error stopping cron scheduler", "error", err)
				return err
			}
			logger.Info("Cron scheduler stopped gracefully")
			return nil
		},
	))
}
