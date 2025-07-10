package queue

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
	mentionService *service.MentionService
	scrapers       []*xscraper.XScraper
	logger         *slog.Logger
	jobQueueClient jobq.JobQueueClient
}

// NewMentionHandler constructs a MentionHandler.
func NewMentionHandler(
	mentionService *service.MentionService,
	scrapers []*xscraper.XScraper,
	logger *slog.Logger,
	jobQueueClient jobq.JobQueueClient,
) *MentionHandler {
	return &MentionHandler{
		mentionService: mentionService,
		scrapers:       scrapers,
		logger:         logger.With("job_handler", "mention"),
		jobQueueClient: jobQueueClient,
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

	// Only process reply mentions - ignore direct mentions
	if !mention.IsReply || mention.InReplyToStatusID == "" {
		w.logger.Info("Skipping non-reply mention",
			"is_reply", mention.IsReply,
			"in_reply_to_status_id", mention.InReplyToStatusID,
			"text", mention.Text)

		return nil
	}

	// Use InReplyToStatusID as the thread ID to scrape
	threadID := mention.InReplyToStatusID

	logger := w.logger.With(
		"job_type", j.Type,
		"mention_user_id", mentionUserID,
		"mention_id", mention.RestID,
		"thread_id", threadID,
		"author_screen_name", mention.Author.ScreenName,
	)

	logger.Info("🤖 Processing mention from queue - creating database records first",
		"text", mention.Text,
		"created_at", mention.CreatedAt.Format(time.RFC3339),
	)

	// Create mention record (this will create a pending thread and trigger ThreadScrapeJob)
	_, err := w.mentionService.CreateMention(ctx, mentionUserID, threadID, &mention.RestID, mention.CreatedAt)
	if err != nil {
		logger.Error("Failed to create mention record", "error", err)
		return fmt.Errorf("failed to create mention from URL: %w", err)
	}

	logger.Info("🤖 Created mention and thread records")

	// Step 2: Create and enqueue thread scrape job (consistent with PostThreadScrape)
	threadScrapeJob, err := NewThreadScrapeJob(threadID)
	if err != nil {
		logger.Error("Failed to create thread scrape job", "error", err)
		return fmt.Errorf("failed to create thread scrape job: %w", err)
	}

	scrapeJobID, err := w.jobQueueClient.Enqueue(ctx, threadScrapeJob)
	if err != nil {
		logger.Error("Failed to enqueue thread scrape job", "error", err)
		return fmt.Errorf("failed to enqueue thread scrape job: %w", err)
	}

	logger.Info("🤖 Enqueued thread scrape job", "scrape_job_id", scrapeJobID)

	// Step 3: Create reply tweet job (keep existing logic)
	replyJob, err := NewReplyTweetJob(mention.RestID, mention.Author.ScreenName)
	if err != nil {
		logger.Error("Failed to create reply tweet job", "error", err)
		return fmt.Errorf("create reply tweet job: %w", err)
	}

	replyJobID, err := w.jobQueueClient.Enqueue(ctx, replyJob)
	if err != nil {
		logger.Error("Failed to enqueue reply tweet job", "error", err)
		return fmt.Errorf("enqueue reply tweet job: %w", err)
	}

	logger.Info("🤖 Mention processed successfully with unified architecture",
		"mention_tweet_id", mention.RestID,
		"thread_id", threadID,
		"scrape_job_id", scrapeJobID,
		"reply_job_id", replyJobID,
		"processing_time", time.Since(mention.CreatedAt),
	)
	return nil
}
