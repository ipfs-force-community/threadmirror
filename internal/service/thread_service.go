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
	"github.com/ipfs-force-community/threadmirror/internal/model"
	wrRedis "github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs/go-cid"
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
	Status  string        `json:"status"`
	Version int           `json:"version"`
	Author  *ThreadAuthor `json:"author,omitempty"`
}

type ThreadRepoInterface interface {
	GetThreadByID(ctx context.Context, id string) (*model.Thread, error)
	CreateThread(ctx context.Context, thread *model.Thread) error
	UpdateThread(ctx context.Context, thread *model.Thread) error
	GetTweetsByIDs(ctx context.Context, ids []string) (map[string]*model.Thread, error)
	UpdateThreadStatus(ctx context.Context, threadID string, status model.ThreadStatus, version int) error
	GetStuckScrapingThreadsForRetry(ctx context.Context, stuckDuration time.Duration, maxRetries int) ([]*model.Thread, error)
	GetOldPendingThreadsForRetry(ctx context.Context, pendingDuration time.Duration, maxRetries int) ([]*model.Thread, error)
	GetFailedThreadsForRetry(ctx context.Context, retryDelay time.Duration, maxRetries int) ([]*model.Thread, error)
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
	threadRepo ThreadRepoInterface
	storage    ipfs.Storage
	cache      cache.CacheInterface[TweetSlice]
	llm        llm.Model
	logger     *slog.Logger
}

func NewThreadService(threadRepo ThreadRepoInterface, storage ipfs.Storage, llmModel llm.Model, redisClientWrapper *wrRedis.Client, logger *slog.Logger) *ThreadService {
	redisStore := redis_store.NewRedis(redisClientWrapper.Client)
	cacheManager := cache.New[TweetSlice](redisStore)
	return &ThreadService{threadRepo: threadRepo, storage: storage, cache: cacheManager, llm: llmModel, logger: logger}
}

func (s *ThreadService) GetThreadByID(ctx context.Context, id string) (*ThreadDetail, error) {
	thread, err := s.threadRepo.GetThreadByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get thread: %w", err)
	}

	// Load tweets from IPFS (only if completed and has CID)
	var tweets []*xscraper.Tweet
	if thread.Status == model.ThreadStatusCompleted && thread.CID != "" {
		tweets, err = s.loadTweetsFromIPFS(ctx, thread.CID)
		if err != nil {
			return nil, fmt.Errorf("load from ipfs %s: %w", thread.CID, err)
		}
	}

	// Build author info if available
	var author *ThreadAuthor
	if thread.AuthorID != "" {
		author = &ThreadAuthor{
			ID:              thread.AuthorID,
			Name:            thread.AuthorName,
			ScreenName:      thread.AuthorScreenName,
			ProfileImageURL: thread.AuthorProfileImageURL,
		}
	}

	return &ThreadDetail{
		ID:             thread.ID,
		CID:            thread.CID,
		ContentPreview: thread.Summary,
		NumTweets:      thread.NumTweets,
		Tweets:         tweets,
		CreatedAt:      thread.CreatedAt,
		Status:         string(thread.Status),
		Version:        thread.Version,
		Author:         author,
	}, nil
}

// loadTweetsFromIPFS loads tweets from IPFS using the CID
func (s *ThreadService) loadTweetsFromIPFS(ctx context.Context, cidStr string) ([]*xscraper.Tweet, error) {
	if cidStr == "" {
		return nil, errutil.ErrNotFound
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
func (s *ThreadService) UpdateThreadStatus(ctx context.Context, threadID string, status model.ThreadStatus, version int) error {
	// Attempt optimistic update
	err := s.threadRepo.UpdateThreadStatus(ctx, threadID, status, version)
	if err != nil {
		return fmt.Errorf("failed to update thread status: %w", err)
	}

	s.logger.Info("thread status updated", "threadID", threadID, "to", status, "version", version+1)
	return nil
}

// GetStuckScrapingThreadsForRetry gets threads that have been in 'scraping' status for too long and increments their retry count
func (s *ThreadService) GetStuckScrapingThreadsForRetry(ctx context.Context, stuckDuration time.Duration, maxRetries int) ([]*model.Thread, error) {
	return s.threadRepo.GetStuckScrapingThreadsForRetry(ctx, stuckDuration, maxRetries)
}

// GetOldPendingThreadsForRetry gets threads that have been in 'pending' status for too long and increments their retry count
func (s *ThreadService) GetOldPendingThreadsForRetry(ctx context.Context, pendingDuration time.Duration, maxRetries int) ([]*model.Thread, error) {
	return s.threadRepo.GetOldPendingThreadsForRetry(ctx, pendingDuration, maxRetries)
}

// GetFailedThreadsForRetry gets failed threads that can be retried and increments their retry count
func (s *ThreadService) GetFailedThreadsForRetry(ctx context.Context, retryDelay time.Duration, maxRetries int) ([]*model.Thread, error) {
	return s.threadRepo.GetFailedThreadsForRetry(ctx, retryDelay, maxRetries)
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

	// Get thread to update (for basic info, not for version)
	thread, err := s.threadRepo.GetThreadByID(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}

	// Set the version passed by caller for optimistic locking
	thread.Version = version

	// Update thread with scraped data
	thread.Summary = summary
	thread.CID = cid.String()
	thread.NumTweets = len(tweets)
	thread.Status = model.ThreadStatusCompleted

	// Set author information from last tweet (which is the thread starter)
	author := tweets[len(tweets)-1].Author
	if author != nil {
		thread.AuthorID = author.RestID
		thread.AuthorName = author.Name
		thread.AuthorScreenName = author.ScreenName
		thread.AuthorProfileImageURL = author.ProfileImageURL
	}

	// Update in database with optimistic locking
	err = s.threadRepo.UpdateThread(ctx, thread)
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
