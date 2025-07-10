package cron

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ipfs-force-community/threadmirror/internal/task/queue"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

// MentionCheckHandler handles checking for new mentions and processing them
type MentionCheckHandler struct {
	logger         *slog.Logger
	scrapers       []*xscraper.XScraper
	jobQueueClient jobq.JobQueueClient

	// Lower-cased prefix of author screen names to exclude from processing
	excludeMentionAuthorPrefixLower string

	// Username to monitor for mentions
	mentionUsername string
}

// MentionCheckConfig holds configuration for the mention check handler
type MentionCheckConfig struct {
	ExcludeMentionAuthorPrefix string `mapstructure:"exclude_mention_author_prefix"`
	MentionUsername            string `mapstructure:"mention_username"`
}

// NewMentionCheckHandler creates a new mention check handler
func NewMentionCheckHandler(
	logger *slog.Logger,
	scrapers []*xscraper.XScraper,
	jobQueueClient jobq.JobQueueClient,
	config MentionCheckConfig,
) *MentionCheckHandler {
	// If mentionUsername is empty, use the first scraper's username as fallback
	mentionUsername := config.MentionUsername
	if mentionUsername == "" && len(scrapers) > 0 {
		mentionUsername = scrapers[0].LoginOpts.Username
	}

	return &MentionCheckHandler{
		logger:                          logger.With("cron_handler", "mention_check"),
		scrapers:                        scrapers,
		jobQueueClient:                  jobQueueClient,
		excludeMentionAuthorPrefixLower: strings.ToLower(config.ExcludeMentionAuthorPrefix),
		mentionUsername:                 mentionUsername,
	}
}

// Execute implements the cron task handler interface
func (h *MentionCheckHandler) Execute(ctx context.Context) error {
	h.logger.Info("Checking for new mentions",
		"mention_username", h.mentionUsername,
		"exclude_prefix", h.excludeMentionAuthorPrefixLower,
	)

	pool := xscraper.NewScraperPool(h.scrapers)
	mentions, err := xscraper.TryWithResult(pool, func(sc *xscraper.XScraper) ([]*xscraper.Tweet, error) {
		filter := func(tweet *xscraper.Tweet) bool {
			if h.excludeMentionAuthorPrefixLower == "" {
				return true
			}
			return !strings.HasPrefix(strings.ToLower(tweet.Author.ScreenName), h.excludeMentionAuthorPrefixLower)
		}
		return sc.GetMentionsByScreenName(ctx, h.mentionUsername, filter)
	})

	if err != nil {
		h.logger.Error("Failed to get mentions", "error", err)
		return fmt.Errorf("get mentions: %w", err)
	}

	if len(mentions) == 0 {
		h.logger.Info("No mentions found")
		return nil
	}

	h.logger.Info("Found mentions", "count", len(mentions))

	// Enqueue each mention as a job
	var lastErr error
	enqueuedCount := 0
	for _, mention := range mentions {
		if err := h.enqueueMentionJob(ctx, mention); err != nil {
			h.logger.Error("Failed to enqueue mention job",
				"mention_id", mention.ID,
				"author", mention.Author.ScreenName,
				"error", err,
			)
			lastErr = err
			continue
		}
		enqueuedCount++
	}

	h.logger.Info("Mention check completed",
		"total_mentions", len(mentions),
		"enqueued_successfully", enqueuedCount,
	)

	// Return the last error if any occurred, but continue processing other mentions
	if lastErr != nil {
		return fmt.Errorf("failed to enqueue some mention jobs (last error): %w", lastErr)
	}

	return nil
}

// enqueueMentionJob enqueues a mention for processing as a job
func (h *MentionCheckHandler) enqueueMentionJob(ctx context.Context, mention *xscraper.Tweet) error {
	job, err := queue.NewMentionJob(mention)
	if err != nil {
		h.logger.Error("Failed to create mention job", "error", err)
		return err
	}
	jobID, err := h.jobQueueClient.Enqueue(ctx, job)
	if err != nil {
		h.logger.Error("Failed to enqueue mention job", "error", err)
		return err
	}
	h.logger.Info("Enqueued mention job", "mention_id", mention.ID, "author", mention.Author.ScreenName, "job_id", jobID)
	return nil
}
