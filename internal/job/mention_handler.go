package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

const TypeProcessMention = "process_mention"

type MentionPayload struct {
	Tweet *xscraper.Tweet `json:"tweet"`
}

type MentionHandler struct {
	processedMarkService *service.ProcessedMarkService
	mentionService       *service.MentionService
	scraper              *xscraper.XScraper
	logger               *slog.Logger
	jobQueueClient       jobq.JobQueueClient
}

// NewMentionHandler constructs a MentionHandler.
func NewMentionHandler(
	processedMarkService *service.ProcessedMarkService,
	mentionService *service.MentionService,
	scraper *xscraper.XScraper,
	logger *slog.Logger,
	jobQueueClient jobq.JobQueueClient,
) *MentionHandler {
	return &MentionHandler{
		processedMarkService: processedMarkService,
		mentionService:       mentionService,
		scraper:              scraper,
		logger:               logger.With("job_handler", "mention"),
		jobQueueClient:       jobQueueClient,
	}
}

// NewMentionJob creates a new generic job for processing a mention with appropriate options.
func NewMentionJob(tweet *xscraper.Tweet) (*jobq.Job, error) {
	payload, err := json.Marshal(MentionPayload{Tweet: tweet})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mention payload: %w", err)
	}

	// Configure job with retry policy and timeout
	return jobq.NewJob(
		TypeProcessMention,
		payload,
	), nil
}

// HandleJob implements the job.JobHandler interface.
func (w *MentionHandler) HandleJob(ctx context.Context, j *jobq.Job) error {
	var payload MentionPayload
	if err := json.Unmarshal(j.Payload, &payload); err != nil {
		// Use SkipRetry for unmarshal errors as they won't resolve with retries
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	mention := payload.Tweet
	if mention == nil {
		// Use SkipRetry for nil tweet as this won't resolve with retries
		return fmt.Errorf("tweet is nil")
	}

	// Validate required fields
	if mention.Author == nil {
		return fmt.Errorf("tweet author is nil")
	}

	mentionUserID := mention.Author.RestID
	if mentionUserID == "" {
		return fmt.Errorf("mention user ID is empty")
	}

	log := w.logger.With(
		"job_type", j.Type,
		"mention_user_id", mentionUserID,
		"tweet_id", mention.RestID,
		"author_screen_name", mention.Author.ScreenName,
	)

	// Check if already processed (idempotency check)
	processed, err := w.processedMarkService.IsProcessed(ctx, mention.RestID, TypeProcessMention)
	if err != nil {
		log.Error("Failed to check if mention is processed", "error", err)
		// This is a transient error, allow retries
		return fmt.Errorf("failed to check if mention is processed: %w", err)
	}

	if processed {
		return nil
	}

	log.Info("🤖 Processing mention from queue",
		"text", mention.Text,
		"created_at", mention.CreatedAt.Format(time.RFC3339),
	)

	tweets, err := w.scraper.GetTweets(ctx, mention.RestID)
	if err != nil {
		log.Error("Failed to get tweets", "error", err)
		return fmt.Errorf("failed to get tweets: %w", err)
	}

	if len(tweets) < 2 {
		return nil
	}

	// Create mention from tweets
	mentionSummary, err := w.mentionService.CreateMention(ctx, &service.CreateMentionRequest{
		Tweets: tweets,
	})
	if err != nil {
		// This could be transient (network issues, db issues), allow retries
		return fmt.Errorf("failed to create mention from tweets: %w", err)
	}

	replyJob, err := NewReplyTweetJob(mentionSummary.ID)
	if err != nil {
		return fmt.Errorf("create reply tweet job: %w", err)
	}
	jobID, err := w.jobQueueClient.Enqueue(ctx, replyJob)
	if err != nil {
		return fmt.Errorf("enqueue reply tweet job: %w", err)
	}

	// Mark as processed to prevent duplicate work
	if err := w.processedMarkService.MarkProcessed(ctx, mention.RestID, TypeProcessMention); err != nil {
		log.Error("Failed to mark mention as processed",
			"error", err,
			"mention_id", mentionSummary.ID,
		)
		// This is serious but the mention was created, so we should retry the marking
		return fmt.Errorf("failed to mark mention as processed: %w", err)
	}

	log.Info("🤖 Mention processed successfully",
		"mention_id", mentionSummary.ID,
		"reply_tweet_job_id", jobID,
		"processing_time", time.Since(mention.CreatedAt),
	)
	return nil
}
