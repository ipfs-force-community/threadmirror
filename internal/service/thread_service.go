package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/google/uuid"
	"github.com/ipfs-force-community/threadmirror/internal/sqlc_generated"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	dbsql "github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs/go-cid"
	"github.com/jackc/pgx/v5"
	"github.com/tmc/langchaingo/llms"
)

type ThreadDetail struct {
	ID             string            `json:"id"`
	CID            string            `json:"cid"`
	ContentPreview string            `json:"content_preview"`
	NumTweets      int               `json:"numTweets"`
	Tweets         []*xscraper.Tweet `json:"tweets,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`

	// New fields for status and author
	Status     string        `json:"status"`
	RetryCount int           `json:"retry_count"`
	Version    int           `json:"version"`
	Author     *ThreadAuthor `json:"author,omitempty"`
}

// TweetSlice is a helper type that implements encoding.BinaryMarshaler and
// encoding.BinaryUnmarshaler so that it can be stored directly in gocache / redis.
type TweetSlice []*xscraper.Tweet

// MarshalBinary encodes the slice as JSON.
func (ts TweetSlice) MarshalBinary() ([]byte, error) {
	return json.Marshal(ts)
}

// UnmarshalBinary decodes the JSON data into the slice.
func (ts *TweetSlice) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, ts)
}

type ThreadService struct {
	db      *dbsql.DB
	storage ipfs.Storage
	cache   cache.CacheInterface[TweetSlice]
	llm     llm.Model
	logger  *slog.Logger
}

func NewThreadService(db *dbsql.DB, storage ipfs.Storage, llmModel llm.Model, redisClientWrapper *redis.Client, logger *slog.Logger) *ThreadService {
	redisStore := redis_store.NewRedis(redisClientWrapper.Client)
	cacheManager := cache.New[TweetSlice](redisStore)
	return &ThreadService{db: db, storage: storage, cache: cacheManager, llm: llmModel, logger: logger}
}

func (s *ThreadService) GetThreadByID(ctx context.Context, id string) (*ThreadDetail, error) {
	threadID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid thread ID: %w", err)
	}

	thread, err := s.db.QueriesFromContext(ctx).GetThreadByID(ctx, sqlc_generated.GetThreadByIDParams{ThreadID: threadID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrThreadNotFound
		}
		return nil, fmt.Errorf("get thread: %w", err)
	}

	// Load tweets from IPFS (only if completed and has CID)
	var tweets []*xscraper.Tweet
	if thread.Status == "completed" && thread.Cid != "" {
		tweets, err = s.loadTweetsFromIPFS(ctx, thread.Cid)
		if err != nil {
			return nil, fmt.Errorf("load from ipfs %s: %w", thread.Cid, err)
		}
	}

	// Build author info if available
	var author *ThreadAuthor
	if thread.AuthorID != nil && *thread.AuthorID != "" {
		author = &ThreadAuthor{
			ID:              *thread.AuthorID,
			Name:            getStringValue(thread.AuthorName),
			ScreenName:      getStringValue(thread.AuthorScreenName),
			ProfileImageURL: getStringValue(thread.AuthorProfileImageUrl),
		}
	}

	return &ThreadDetail{
		ID:             thread.ID.String(),
		CID:            thread.Cid,
		ContentPreview: thread.Summary,
		NumTweets:      int(thread.NumTweets),
		Tweets:         tweets,
		CreatedAt:      thread.CreatedAt,
		Status:         thread.Status,
		RetryCount:     int(thread.RetryCount),
		Version:        int(thread.Version),
		Author:         author,
	}, nil
}

// loadTweetsFromIPFS loads tweets from IPFS using the CID
func (s *ThreadService) loadTweetsFromIPFS(ctx context.Context, cidStr string) ([]*xscraper.Tweet, error) {
	if cidStr == "" {
		return nil, ErrNotFound
	}

	// First, try cache
	if s.cache != nil {
		key := cacheKeyForThread(cidStr)

		cached, err := s.cache.Get(ctx, key)
		// Successful cache hit
		if err == nil && cached != nil {
			return []*xscraper.Tweet(cached), nil
		}
		// Log errors other than a simple not-found miss
		if err != nil && !errors.Is(err, store.NotFound{}) {
			s.logger.Error("failed to get cache", "error", err)
		}
	}

	// Parse CID from string
	c, err := cid.Parse(cidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CID: %w", err)
	}

	// Get content from IPFS
	reader, err := s.storage.Get(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to get content from IPFS: %w", err)
	}
	defer reader.Close() // nolint:errcheck

	// Read all content
	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Unmarshal tweets
	var tweets []*xscraper.Tweet
	err = json.Unmarshal(buffer.Bytes(), &tweets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tweets: %w", err)
	}

	// Cache the result for future requests
	if s.cache != nil {
		// Permanent cache (no TTL), rely on Redis allkeys-lru eviction
		key := cacheKeyForThread(cidStr)
		err = s.cache.Set(ctx, key, TweetSlice(tweets))
		if err != nil {
			s.logger.Error("failed to set cache", "error", err)
		}
	}
	return tweets, nil
}

// cacheKeyForThread returns a namespaced cache key for storing tweets by CID
func cacheKeyForThread(cid string) string {
	return "thread:" + cid
}

// UpdateThreadStatus updates thread status with optimistic locking
func (s *ThreadService) UpdateThreadStatus(ctx context.Context, threadID string, status string, version int) error {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return fmt.Errorf("invalid thread ID: %w", err)
	}

	err = s.db.QueriesFromContext(ctx).UpdateThreadStatus(ctx, sqlc_generated.UpdateThreadStatusParams{
		ThreadID:       threadUUID,
		Status:         status,
		CurrentVersion: int32(version),
	})
	if err != nil {
		return fmt.Errorf("failed to update thread status: %w", err)
	}

	s.logger.Info("thread status updated", "threadID", threadID, "to", status, "version", version+1)
	return nil
}

// GetStuckScrapingThreadsForRetry gets threads that have been in 'scraping' status for too long and increments their retry count
func (s *ThreadService) GetStuckScrapingThreadsForRetry(ctx context.Context, stuckDuration time.Duration, maxRetries int) ([]sqlc_generated.Thread, error) {
	cutoffTime := time.Now().Add(-stuckDuration)

	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := s.db.QueriesFromContext(ctx).WithTx(tx)

	// Get stuck threads
	threads, err := queries.GetStuckScrapingThreads(ctx, sqlc_generated.GetStuckScrapingThreadsParams{
		CutoffTime: cutoffTime,
		MaxRetries: int32(maxRetries),
	})
	if err != nil {
		return nil, fmt.Errorf("get stuck threads: %w", err)
	}

	if len(threads) == 0 {
		return threads, nil
	}

	// Extract thread IDs
	threadIDs := make([]uuid.UUID, len(threads))
	for i, thread := range threads {
		threadIDs[i] = thread.ID
	}

	// Increment retry count
	err = queries.IncrementThreadRetryCount(ctx, sqlc_generated.IncrementThreadRetryCountParams{ThreadIds: threadIDs})
	if err != nil {
		return nil, fmt.Errorf("increment retry count: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return threads, nil
}

// GetOldPendingThreadsForRetry gets threads that have been in 'pending' status for too long and increments their retry count
func (s *ThreadService) GetOldPendingThreadsForRetry(ctx context.Context, pendingDuration time.Duration, maxRetries int) ([]sqlc_generated.Thread, error) {
	cutoffTime := time.Now().Add(-pendingDuration)

	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := s.db.QueriesFromContext(ctx).WithTx(tx)

	// Get old pending threads
	threads, err := queries.GetOldPendingThreads(ctx, sqlc_generated.GetOldPendingThreadsParams{
		CutoffTime: cutoffTime,
		MaxRetries: int32(maxRetries),
	})
	if err != nil {
		return nil, fmt.Errorf("get old pending threads: %w", err)
	}

	if len(threads) == 0 {
		return threads, nil
	}

	// Extract thread IDs
	threadIDs := make([]uuid.UUID, len(threads))
	for i, thread := range threads {
		threadIDs[i] = thread.ID
	}

	// Increment retry count
	err = queries.IncrementThreadRetryCount(ctx, sqlc_generated.IncrementThreadRetryCountParams{ThreadIds: threadIDs})
	if err != nil {
		return nil, fmt.Errorf("increment retry count: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return threads, nil
}

// GetFailedThreadsForRetry gets failed threads that can be retried and increments their retry count
func (s *ThreadService) GetFailedThreadsForRetry(ctx context.Context, retryDelay time.Duration, maxRetries int) ([]sqlc_generated.Thread, error) {
	cutoffTime := time.Now().Add(-retryDelay)

	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := s.db.QueriesFromContext(ctx).WithTx(tx)

	// Get failed threads for retry
	threads, err := queries.GetFailedThreadsForRetry(ctx, sqlc_generated.GetFailedThreadsForRetryParams{
		CutoffTime: cutoffTime,
		MaxRetries: int32(maxRetries),
	})
	if err != nil {
		return nil, fmt.Errorf("get failed threads: %w", err)
	}

	if len(threads) == 0 {
		return threads, nil
	}

	// Extract thread IDs
	threadIDs := make([]uuid.UUID, len(threads))
	for i, thread := range threads {
		threadIDs[i] = thread.ID
	}

	// Increment retry count
	err = queries.IncrementThreadRetryCount(ctx, sqlc_generated.IncrementThreadRetryCountParams{ThreadIds: threadIDs})
	if err != nil {
		return nil, fmt.Errorf("increment retry count: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return threads, nil
}

// UpdateThreadWithScrapedData updates thread with complete scraped data including summary generation and IPFS upload
func (s *ThreadService) UpdateThreadWithScrapedData(
	ctx context.Context,
	threadID string,
	tweets []*xscraper.Tweet,
	version int,
) error {
	if len(tweets) == 0 {
		return fmt.Errorf("no tweets provided")
	}

	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return fmt.Errorf("invalid thread ID: %w", err)
	}

	// Generate summary using the same logic as MentionService
	summary, err := s.generateTweetsSummary(ctx, tweets)
	if err != nil {
		return fmt.Errorf("failed to generate AI summary: %w", err)
	}

	// Marshal and upload tweets to IPFS
	jsonTweets, err := json.Marshal(tweets)
	if err != nil {
		return fmt.Errorf("failed to marshal tweets: %w", err)
	}

	cid, err := s.storage.Add(ctx, bytes.NewReader(jsonTweets))
	if err != nil {
		return fmt.Errorf("failed to add tweets to IPFS: %w", err)
	}

	// Set author information from last tweet (which is the thread starter)
	var authorID, authorName, authorScreenName, authorProfileImageURL *string
	author := tweets[len(tweets)-1].Author
	if author != nil {
		authorID = &author.RestID
		authorName = &author.Name
		authorScreenName = &author.ScreenName
		authorProfileImageURL = &author.ProfileImageURL
	}

	// Update thread with scraped data using optimistic locking
	err = s.db.QueriesFromContext(ctx).UpdateThreadComplete(ctx, sqlc_generated.UpdateThreadCompleteParams{
		ID:                    threadUUID,
		Summary:               summary,
		Cid:                   cid.String(),
		NumTweets:             int32(len(tweets)),
		Status:                "completed",
		RetryCount:            0, // Reset retry count on successful completion
		ExpectedVersion:       int32(version),
		AuthorID:              authorID,
		AuthorName:            authorName,
		AuthorScreenName:      authorScreenName,
		AuthorProfileImageUrl: authorProfileImageURL,
	})
	if err != nil {
		return fmt.Errorf("failed to update thread: %w", err)
	}

	s.logger.Info("thread updated successfully", "threadID", threadID, "version", version)
	return nil
}

// generateTweetsSummary generates AI summary for tweets using LLM
func (s *ThreadService) generateTweetsSummary(ctx context.Context, tweets []*xscraper.Tweet) (string, error) {
	if len(tweets) == 0 {
		return "Empty thread", nil
	}

	type ToSummarize struct {
		Text   string `json:"text"`
		Author struct {
			Name       string `json:"name"`
			ScreenName string `json:"screen_name"`
		} `json:"author"`
		IsRetweet bool `json:"is_retweet"`
		IsReply   bool `json:"is_reply"`
	}
	toSummarize := make([]ToSummarize, 0, len(tweets))
	for _, tweet := range tweets {
		toSummarize = append(toSummarize, ToSummarize{
			Text: tweet.Text,
			Author: struct {
				Name       string `json:"name"`
				ScreenName string `json:"screen_name"`
			}{
				Name:       tweet.Author.Name,
				ScreenName: tweet.Author.ScreenName,
			},
			IsRetweet: tweet.IsRetweet,
			IsReply:   tweet.IsReply,
		})
	}

	jsonTweets, err := json.Marshal(toSummarize)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tweets: %w", err)
	}

	// Create prompt for AI summarization
	prompt := fmt.Sprintf(`Please analyze the following JSON data containing Twitter/X posts and provide a concise summary (maximum 200 characters) in Chinese. 

The JSON contains an array of tweet objects, each with fields like "text", "author", etc. Focus on the main content and key themes from the "text" fields.

JSON Data:
%s

Please provide a Chinese summary:`, jsonTweets)

	// Generate summary using LLM
	summary, err := llms.GenerateFromSinglePrompt(ctx, s.llm, prompt,
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}
	summary = strings.TrimSpace(strings.ToValidUTF8(summary, ""))

	// Ensure summary is not too long
	if len([]rune(summary)) > 200 {
		runes := []rune(summary)
		summary = string(runes[:200]) + "..."
	}
	// Filter out invalid UTF-8 characters
	return strings.TrimSpace(summary), nil
}
