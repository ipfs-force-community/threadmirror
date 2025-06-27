package bot

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

// TwitterBot represents a Twitter bot that responds to mentions
type TwitterBot struct {
	// Bot credentials and settings
	username         string
	email            string
	checkInterval    time.Duration
	maxMentionsCheck int

	scraper                 *xscraper.XScraper
	botCookieService        *service.BotCookieService
	processedMentionService *service.ProcessedMentionService
	postService             *service.PostService
	logger                  *slog.Logger

	// Control channels
	stopCh  chan struct{}
	stopped chan struct{}
}

// NewTwitterBot creates a new Twitter bot instance
func NewTwitterBot(
	username, password, email string,
	checkInterval time.Duration,
	maxMentionsCheck int,
	processedMentionService *service.ProcessedMentionService,
	botCookieService *service.BotCookieService,
	postService *service.PostService,
	logger *slog.Logger,
) *TwitterBot {
	// Create login options for xscraper
	loginOpts := xscraper.LoginOptions{
		LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
			return botCookieService.LoadCookies(ctx, email, username)
		},
		SaveCookies: func(ctx context.Context, cookies []*http.Cookie) error {
			return botCookieService.SaveCookies(ctx, email, username, cookies)
		},
		Username: username,
		Password: password,
		Email:    email,
	}

	// Create xscraper instance
	scraper := xscraper.New(loginOpts, logger)

	return &TwitterBot{
		username:                username,
		email:                   email,
		checkInterval:           checkInterval,
		maxMentionsCheck:        maxMentionsCheck,
		scraper:                 scraper,
		botCookieService:        botCookieService,
		processedMentionService: processedMentionService,
		postService:             postService,
		logger:                  logger,
		stopCh:                  make(chan struct{}),
		stopped:                 make(chan struct{}),
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

	// Process each mention
	for _, mention := range mentions {
		if err := tb.processMention(ctx, mention); err != nil {
			tb.logger.Error("Failed to process mention",
				"mention_id", mention.ID,
				"author", mention.Author.ScreenName,
				"error", err,
			)
		}
	}

	return nil
}

// processMention processes a single mention (log only)
func (tb *TwitterBot) processMention(ctx context.Context, mention *xscraper.Tweet) error {
	// Use the author's user ID (the user who mentioned the bot) + tweet ID to track processing
	mentionUserID := mention.Author.ID

	logger := tb.logger.With("mention_user_id", mentionUserID, "tweet_id", mention.ID)

	// Check if we've already processed this mention from this author
	isProcessed, err := tb.processedMentionService.IsProcessed(ctx, mentionUserID, mention.ID)
	if err != nil {
		logger.Error("Failed to check if mention is processed", "error", err)
		return fmt.Errorf("failed to check if mention is processed: %w", err)
	}

	if isProcessed {
		logger.Debug("Mention already processed")
		return nil
	}

	// Log the detected mention
	logger.Info("🤖 Detected new mention",
		"text", mention.Text,
		"created_at", mention.CreatedAt.Format(time.RFC3339),
	)

	post, err := tb.postService.CreatePost(ctx, mentionUserID, &service.CreatePostRequest{
		Tweets: []*xscraper.Tweet{mention},
	})

	if err != nil {
		logger.Error("Failed to create post", "error", err)
		return fmt.Errorf("failed to create post: %w", err)
	}

	logger.Info("🤖 Created post", "post_id", post.ID)

	// Mark as processed to avoid duplicate logging
	if err := tb.processedMentionService.MarkProcessed(ctx, mentionUserID, mention.ID); err != nil {
		logger.Error("Failed to mark mention as processed", "error", err)
		return fmt.Errorf("failed to mark mention as processed: %w", err)
	}

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
