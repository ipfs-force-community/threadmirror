package bot

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/job"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

// TwitterBot represents a Twitter bot that responds to mentions
type TwitterBot struct {
	// Bot credentials and settings
	username         string
	email            string
	checkInterval    time.Duration
	maxMentionsCheck int

	scraper        *xscraper.XScraper
	jobQueueClient jobq.JobQueueClient
	logger         *slog.Logger

	// Control channels
	stopCh  chan struct{}
	stopped chan struct{}
}

// NewTwitterBot creates a new Twitter bot instance
func NewTwitterBot(
	username string,
	email string,
	scraper *xscraper.XScraper,
	checkInterval time.Duration,
	maxMentionsCheck int,
	jobQueueClient jobq.JobQueueClient,
	logger *slog.Logger,
) *TwitterBot {
	return &TwitterBot{
		username:         username,
		email:            email,
		checkInterval:    checkInterval,
		maxMentionsCheck: maxMentionsCheck,
		scraper:          scraper,
		jobQueueClient:   jobQueueClient,
		logger:           logger,
		stopCh:           make(chan struct{}),
		stopped:          make(chan struct{}),
	}
}

// randomizedInterval returns the base interval with ±30% random variation
func (tb *TwitterBot) randomizedInterval() time.Duration {
	// 30% jitter range
	jitterRange := float64(tb.checkInterval) * 0.3

	// Generate random jitter between -30% and +30%
	jitter := time.Duration((rand.Float64() - 0.5) * 2 * jitterRange)
	interval := tb.checkInterval + jitter

	// Ensure minimum of 30 seconds
	interval = max(interval, 30*time.Second)

	tb.logger.Debug("Randomized interval",
		"base", tb.checkInterval,
		"jitter", jitter,
		"final", interval)

	return interval
}

// Start starts the Twitter bot
func (tb *TwitterBot) Start(ctx context.Context) error {
	tb.logger.Info("Starting Twitter bot",
		"username", tb.username,
		"check_interval", tb.checkInterval,
		"max_mentions", tb.maxMentionsCheck,
	)

	go tb.run(context.Background())
	return nil
}

// Stop stops the Twitter bot
func (tb *TwitterBot) Stop(ctx context.Context) error {
	tb.logger.Info("Stopping Twitter bot")
	close(tb.stopCh)
	// Wait for shutdown with context timeout
	select {
	case <-tb.stopped:
		tb.logger.Info("Twitter bot stopped")
		return nil
	case <-ctx.Done():
		tb.logger.Warn("Context cancelled while waiting for bot to stop", "error", ctx.Err())
		return fmt.Errorf("stop operation cancelled: %w", ctx.Err())
	}
}

// run is the main bot loop with randomized intervals
func (tb *TwitterBot) run(ctx context.Context) {
	defer close(tb.stopped)

	for {
		interval := tb.randomizedInterval()
		timer := time.NewTimer(interval)

		select {
		case <-ctx.Done():
			timer.Stop()
			tb.logger.Info("Context cancelled, stopping bot")
			return
		case <-tb.stopCh:
			timer.Stop()
			tb.logger.Info("Stop signal received, stopping bot")
			return
		case <-timer.C:
			if err := tb.checkMentions(ctx); err != nil {
				tb.logger.Error("Failed to check mentions", "error", err)
			}
		}
	}
}

// checkMentions checks for new mentions and responds to them
func (tb *TwitterBot) checkMentions(ctx context.Context) error {
	tb.logger.Debug("Checking for new mentions")

	// Get recent mentions
	mentions, err := tb.scraper.GetMentions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get mentions: %w", err)
	}

	tb.logger.Debug("Found mentions", "count", len(mentions))

	// Enqueue each mention as a job
	for _, mention := range mentions {
		if err := tb.enqueueMentionJob(ctx, mention); err != nil {
			tb.logger.Error("Failed to enqueue mention job",
				"mention_id", mention.ID,
				"author", mention.Author.ScreenName,
				"error", err,
			)
		}
	}

	return nil
}

// enqueueMentionJob enqueues a mention for processing as a job
func (tb *TwitterBot) enqueueMentionJob(ctx context.Context, mention *xscraper.Tweet) error {
	job, err := job.NewMentionJob(mention)
	if err != nil {
		tb.logger.Error("Failed to create mention job", "error", err)
		return err
	}
	_, err = tb.jobQueueClient.Enqueue(ctx, job)
	if err != nil {
		tb.logger.Error("Failed to enqueue mention job", "error", err)
		return err
	}
	tb.logger.Info("Enqueued mention job", "mention_id", mention.ID, "author", mention.Author.ScreenName)
	return nil
}

// GetStats returns bot statistics
func (tb *TwitterBot) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":        true, // Bot is always enabled now
		"username":       tb.username,
		"check_interval": tb.checkInterval.String(),
		"randomized":     true,       // Intervals are randomized with ±30% jitter
		"storage_type":   "database", // Now using database storage for both processed mentions and cookies
	}
}
