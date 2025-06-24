package bot

import (
	"context"
	"fmt"
	"log/slog"
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
		logger:                  logger,
		stopCh:                  make(chan struct{}),
		stopped:                 make(chan struct{}),
	}
}

// Start starts the Twitter bot
func (tb *TwitterBot) Start(ctx context.Context) error {
	tb.logger.Info("Starting Twitter bot",
		"username", tb.username,
		"check_interval", tb.checkInterval,
		"max_mentions", tb.maxMentionsCheck,
	)

	go tb.run(ctx)
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

// run is the main bot loop
func (tb *TwitterBot) run(ctx context.Context) {
	defer close(tb.stopped)

	ticker := time.NewTicker(tb.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tb.logger.Info("Context cancelled, stopping bot")
			return
		case <-tb.stopCh:
			tb.logger.Info("Stop signal received, stopping bot")
			return
		case <-ticker.C:
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
	authorUserID := mention.Author.ID

	// Check if we've already processed this mention from this author
	isProcessed, err := tb.processedMentionService.IsProcessed(ctx, authorUserID, mention.ID)
	if err != nil {
		tb.logger.Error("Failed to check if mention is processed",
			"author_user_id", authorUserID,
			"tweet_id", mention.ID,
			"error", err)
		return fmt.Errorf("failed to check if mention is processed: %w", err)
	}

	if isProcessed {
		tb.logger.Debug("Mention already processed",
			"author_user_id", authorUserID,
			"tweet_id", mention.ID)
		return nil
	}

	// Log the detected mention
	tb.logger.Info("🤖 Detected new mention",
		"author_user_id", authorUserID,
		"tweet_id", mention.ID,
		"author", mention.Author.ScreenName,
		"text", mention.Text,
		"created_at", mention.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	// Mark as processed to avoid duplicate logging
	if err := tb.processedMentionService.MarkProcessed(ctx, authorUserID, mention.ID); err != nil {
		tb.logger.Error("Failed to mark mention as processed",
			"author_user_id", authorUserID,
			"tweet_id", mention.ID,
			"error", err)
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
		"storage_type":   "database", // Now using database storage for both processed mentions and cookies
	}
}
