package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/task/queue"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
)

// ThreadStatusCleanupHandler handles cleanup of threads with problematic statuses
type ThreadStatusCleanupHandler struct {
	logger         *slog.Logger
	threadService  *service.ThreadService
	jobQueueClient jobq.JobQueueClient

	// Configuration
	scrapingTimeout time.Duration // How long before a 'scraping' thread is considered stuck
	pendingTimeout  time.Duration // How long before a 'pending' thread is considered stale
	retryDelay      time.Duration // Minimum delay before retrying failed threads
	maxRetries      int           // Maximum retry attempts (based on retry_count field)
}

// ThreadStatusCleanupConfig holds configuration for the cleanup handler
type ThreadStatusCleanupConfig struct {
	ScrapingTimeoutMinutes int `mapstructure:"scraping_timeout_minutes" default:"30"`
	PendingTimeoutMinutes  int `mapstructure:"pending_timeout_minutes" default:"60"`
	RetryDelayMinutes      int `mapstructure:"retry_delay_minutes" default:"15"`
	MaxRetries             int `mapstructure:"max_retries" default:"3"`
}

// NewThreadStatusCleanupHandler creates a new thread status cleanup handler
func NewThreadStatusCleanupHandler(
	logger *slog.Logger,
	threadService *service.ThreadService,
	jobQueueClient jobq.JobQueueClient,
	config ThreadStatusCleanupConfig,
) *ThreadStatusCleanupHandler {
	// Apply defaults if not set
	if config.ScrapingTimeoutMinutes <= 0 {
		config.ScrapingTimeoutMinutes = 30
	}
	if config.PendingTimeoutMinutes <= 0 {
		config.PendingTimeoutMinutes = 60
	}
	if config.RetryDelayMinutes <= 0 {
		config.RetryDelayMinutes = 15
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}

	return &ThreadStatusCleanupHandler{
		logger:          logger.With("cron_handler", "thread_status_cleanup"),
		threadService:   threadService,
		jobQueueClient:  jobQueueClient,
		scrapingTimeout: time.Duration(config.ScrapingTimeoutMinutes) * time.Minute,
		pendingTimeout:  time.Duration(config.PendingTimeoutMinutes) * time.Minute,
		retryDelay:      time.Duration(config.RetryDelayMinutes) * time.Minute,
		maxRetries:      config.MaxRetries,
	}
}

// Execute implements common.CronTaskHandler
func (h *ThreadStatusCleanupHandler) Execute(ctx context.Context) error {
	h.logger.Info("Starting thread status cleanup task",
		"scraping_timeout", h.scrapingTimeout,
		"pending_timeout", h.pendingTimeout,
		"retry_delay", h.retryDelay,
		"max_retries", h.maxRetries,
	)

	// Handle stuck scraping threads
	stuckThreads, err := h.handleStuckScrapingThreads(ctx)
	if err != nil {
		h.logger.Error("Failed to handle stuck scraping threads", "error", err)
		return fmt.Errorf("handle stuck scraping threads: %w", err)
	}

	// Handle old pending threads
	oldPendingThreads, err := h.handleOldPendingThreads(ctx)
	if err != nil {
		h.logger.Error("Failed to handle old pending threads", "error", err)
		return fmt.Errorf("handle old pending threads: %w", err)
	}

	// Handle failed threads for retry
	retriedThreads, err := h.handleFailedThreadsRetry(ctx)
	if err != nil {
		h.logger.Error("Failed to handle failed threads retry", "error", err)
		return fmt.Errorf("handle failed threads retry: %w", err)
	}

	h.logger.Info("Thread status cleanup completed",
		"stuck_threads_reset", stuckThreads,
		"old_pending_threads_reset", oldPendingThreads,
		"failed_threads_retried", retriedThreads,
	)

	return nil
}

// handleStuckScrapingThreads finds threads stuck in 'scraping' status and resets them to 'pending'
func (h *ThreadStatusCleanupHandler) handleStuckScrapingThreads(ctx context.Context) (int, error) {
	threads, err := h.threadService.GetStuckScrapingThreadsForRetry(ctx, h.scrapingTimeout, h.maxRetries)
	if err != nil {
		return 0, fmt.Errorf("get stuck scraping threads: %w", err)
	}

	if len(threads) == 0 {
		h.logger.Debug("No stuck scraping threads found")
		return 0, nil
	}

	h.logger.Info("Found stuck scraping threads", "count", len(threads))

	resetCount := 0
	for _, thread := range threads {
		logger := h.logger.With(
			"thread_id", thread.ID,
			"stuck_duration", time.Since(thread.UpdatedAt),
			"retry_count", thread.RetryCount, // Note: retry_count was already incremented by GetStuckScrapingThreadsForRetry
		)

		// Note: retry_count was already incremented when fetching threads
		// DB query already filters threads with retry_count < maxRetries

		// Reset status to pending to allow reprocessing
		err := h.threadService.UpdateThreadStatus(ctx, thread.ID, model.ThreadStatusPending, thread.Version)
		if err != nil {
			logger.Error("Failed to reset stuck thread status", "error", err)
			continue
		}

		logger.Info("Reset stuck thread to pending status")
		resetCount++

		// Re-enqueue thread scrape job
		job, err := queue.NewThreadScrapeJob(thread.ID)
		if err != nil {
			logger.Error("Failed to create thread scrape job", "error", err)
			continue
		}

		_, err = h.jobQueueClient.Enqueue(ctx, job)
		if err != nil {
			logger.Error("Failed to enqueue thread scrape job", "error", err)
			continue
		}

		logger.Info("Re-enqueued thread scrape job")
	}

	return resetCount, nil
}

// handleOldPendingThreads finds threads that have been pending for too long and requeue them
func (h *ThreadStatusCleanupHandler) handleOldPendingThreads(ctx context.Context) (int, error) {
	threads, err := h.threadService.GetOldPendingThreadsForRetry(ctx, h.pendingTimeout, h.maxRetries)
	if err != nil {
		return 0, fmt.Errorf("get old pending threads: %w", err)
	}

	if len(threads) == 0 {
		h.logger.Debug("No old pending threads found")
		return 0, nil
	}

	h.logger.Info("Found old pending threads", "count", len(threads))

	requeuedCount := 0
	for _, thread := range threads {
		logger := h.logger.With(
			"thread_id", thread.ID,
			"pending_duration", time.Since(thread.CreatedAt),
			"retry_count", thread.RetryCount, // Note: retry_count was already incremented by GetOldPendingThreadsForRetry
		)

		// Re-enqueue thread scrape job
		job, err := queue.NewThreadScrapeJob(thread.ID)
		if err != nil {
			logger.Error("Failed to create thread scrape job for old pending thread", "error", err)
			continue
		}

		_, err = h.jobQueueClient.Enqueue(ctx, job)
		if err != nil {
			logger.Error("Failed to enqueue thread scrape job for old pending thread", "error", err)
			continue
		}

		logger.Info("Re-enqueued old pending thread")
		requeuedCount++
	}

	return requeuedCount, nil
}

// handleFailedThreadsRetry finds failed threads that can be retried
func (h *ThreadStatusCleanupHandler) handleFailedThreadsRetry(ctx context.Context) (int, error) {
	threads, err := h.threadService.GetFailedThreadsForRetry(ctx, h.retryDelay, h.maxRetries)
	if err != nil {
		return 0, fmt.Errorf("get failed threads for retry: %w", err)
	}

	if len(threads) == 0 {
		h.logger.Debug("No failed threads found for retry")
		return 0, nil
	}

	h.logger.Info("Found failed threads for retry", "count", len(threads))

	retriedCount := 0
	for _, thread := range threads {
		logger := h.logger.With(
			"thread_id", thread.ID,
			"retry_count", thread.RetryCount, // Note: retry_count was already incremented by GetFailedThreadsForRetry
			"time_since_failure", time.Since(thread.UpdatedAt),
		)

		// Note: retry_count was already incremented when fetching threads
		// DB query already filters threads with retry_count < maxRetries

		// Reset status to pending for retry
		err := h.threadService.UpdateThreadStatus(ctx, thread.ID, model.ThreadStatusPending, thread.Version)
		if err != nil {
			logger.Error("Failed to reset failed thread status for retry", "error", err)
			continue
		}

		logger.Info("Reset failed thread to pending for retry")

		// Re-enqueue thread scrape job
		job, err := queue.NewThreadScrapeJob(thread.ID)
		if err != nil {
			logger.Error("Failed to create thread scrape job for retry", "error", err)
			continue
		}

		_, err = h.jobQueueClient.Enqueue(ctx, job)
		if err != nil {
			logger.Error("Failed to enqueue thread scrape job for retry", "error", err)
			continue
		}

		logger.Info("Re-enqueued failed thread for retry")
		retriedCount++
	}

	return retriedCount, nil
}
