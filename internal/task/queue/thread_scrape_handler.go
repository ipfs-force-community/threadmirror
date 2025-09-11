package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/samber/lo"
)

const TypeThreadScrape = "thread_scrape"

type ThreadScrapePayload struct {
	TweetID string `json:"tweet_id"`
}

type ThreadScrapeHandler struct {
	mentionService *service.MentionService
	threadService  *service.ThreadService
	scrapers       []*xscraper.XScraper
	logger         *slog.Logger
}

// NewThreadScrapeHandler constructs a ThreadScrapeHandler.
func NewThreadScrapeHandler(
	mentionService *service.MentionService,
	threadService *service.ThreadService,
	scrapers []*xscraper.XScraper,
	logger *slog.Logger,
) *ThreadScrapeHandler {
	return &ThreadScrapeHandler{
		mentionService: mentionService,
		threadService:  threadService,
		scrapers:       scrapers,
		logger:         logger.With("job_handler", "thread_scrape"),
	}
}

// NewThreadScrapeJob creates a new job for scraping a thread.
func NewThreadScrapeJob(tweetID string) (*jobq.Job, error) {
	payload, err := json.Marshal(ThreadScrapePayload{
		TweetID: tweetID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal thread scrape payload: %w", err)
	}

	return jobq.NewJob(TypeThreadScrape, payload), nil
}

// HandleJob implements the job.JobHandler interface.
func (h *ThreadScrapeHandler) HandleJob(ctx context.Context, j *jobq.Job) error {
	var payload ThreadScrapePayload
	if err := json.Unmarshal(j.Payload, &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	if payload.TweetID == "" {
		return fmt.Errorf("tweet ID is empty")
	}

	logger := h.logger.With(
		"job_type", j.Type,
		"tweet_id", payload.TweetID,
	)

	// Check if thread already exists and determine action based on status
	existingThread, err := h.threadService.GetThreadByID(ctx, payload.TweetID)
	if err != nil && !errors.Is(err, service.ErrThreadNotFound) {
		logger.Error("Failed to check existing thread", "error", err)
		return fmt.Errorf("failed to check existing thread: %w", err)
	}

	// Handle different thread statuses
	if existingThread != nil {
		logger.Info("Thread exists with status", "thread_id", existingThread.ID, "status", existingThread.Status)

		switch existingThread.Status {
		case "completed":
			// Thread already completed, mark as processed and skip
			logger.Info("Thread already completed, skipping scrape", "thread_id", existingThread.ID)
			return nil

		case "scraping":
			// Thread currently being scraped (concurrent job), skip to avoid duplicate work
			logger.Info("Thread already being scraped, skipping", "thread_id", existingThread.ID)
			return nil

		case "pending", "failed":
			// Thread is pending (user submitted URL) or failed (needs retry) - continue with scraping
			logger.Info("Thread ready for scraping", "thread_id", existingThread.ID, "status", existingThread.Status)
			// Continue with scraping logic below

		default:
			// Unknown status, log warning but continue
			logger.Warn("Unknown thread status, proceeding with scraping", "thread_id", existingThread.ID, "status", existingThread.Status)
		}
	}

	logger.Info("🤖 Starting thread scraping job")

	// Update thread status to scraping with optimistic locking
	err = h.threadService.UpdateThreadStatus(ctx, payload.TweetID, "scraping", existingThread.Version)
	if err != nil {
		return fmt.Errorf("failed to update thread status to scraping: %w", err)
	}

	// Use xscraper to get complete thread
	tweets, err := h.getCompleteThreadFromTweetID(ctx, payload.TweetID)
	if err != nil {
		logger.Error("Failed to get complete thread", "error", err)
		if errors.Is(err, service.ErrThreadNotFound) {
			return fmt.Errorf("thread not found: %w", err)
		}
		return fmt.Errorf("failed to get complete thread: %w", err)
	}

	if len(tweets) == 0 {
		logger.Error("No tweets found")
		return fmt.Errorf("no tweets found for thread %s", payload.TweetID)
	}

	// Filter out empty RestID tweets (deleted tweets)
	tweets = lo.Filter(tweets, func(tweet *xscraper.Tweet, _ int) bool {
		return tweet.RestID != ""
	})

	if len(tweets) < 1 {
		logger.Error("No valid tweets found after filtering")
		return fmt.Errorf("no valid tweets found for thread %s", payload.TweetID)
	}

	logger.Info("🤖 Successfully scraped tweets", "count", len(tweets))

	// Get fresh thread version for final update
	finalThread, err := h.threadService.GetThreadByID(ctx, payload.TweetID)
	if err != nil {
		return fmt.Errorf("failed to get thread for final update: %w", err)
	}

	// Update the existing thread with scraped data (status, author info, content)
	err = h.threadService.UpdateThreadWithScrapedData(ctx, payload.TweetID, tweets, finalThread.Version)
	if err != nil {
		_ = h.threadService.UpdateThreadStatus(ctx, payload.TweetID, "failed", finalThread.Version)
		logger.Error("Failed to update thread after scraping", "error", err)
		return fmt.Errorf("failed to update thread after scraping: %w", err)
	}

	logger.Info("🤖 Thread updated successfully with scraped data")

	logger.Info("🤖 Thread scrape completed successfully",
		"thread_id", payload.TweetID,
		"tweets_count", len(tweets),
	)

	return nil
}

// getCompleteThreadFromTweetID gets complete thread using xscraper
func (h *ThreadScrapeHandler) getCompleteThreadFromTweetID(ctx context.Context, tweetID string) ([]*xscraper.Tweet, error) {
	if len(h.scrapers) == 0 {
		return nil, errors.New("no scrapers available")
	}

	pool := xscraper.NewScraperPool(h.scrapers)
	tweets, err := xscraper.TryWithResult(pool, func(sc *xscraper.XScraper) ([]*xscraper.Tweet, error) {
		return xscraper.GetCompleteThread(ctx, sc, tweetID, 0)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get tweets: %w", err)
	}

	return tweets, nil
}
