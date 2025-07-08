package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/ipfs-force-community/threadmirror/internal/model"
	wrRedis "github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs/go-cid"
)

type ThreadDetail struct {
	ID             string            `json:"id"`
	CID            string            `json:"cid"`
	ContentPreview string            `json:"content_preview"`
	NumTweets      int               `json:"numTweets"`
	Tweets         []*xscraper.Tweet `json:"tweets,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type ThreadRepoInterface interface {
	GetThreadByID(ctx context.Context, id string) (*model.Thread, error)
	CreateThread(ctx context.Context, thread *model.Thread) error
	GetTweetsByIDs(ctx context.Context, ids []string) (map[string]*model.Thread, error)
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
	logger     *slog.Logger
}

func NewThreadService(threadRepo ThreadRepoInterface, storage ipfs.Storage, redisClientWrapper *wrRedis.Client, logger *slog.Logger) *ThreadService {
	redisStore := redis_store.NewRedis(redisClientWrapper.Client)
	cacheManager := cache.New[TweetSlice](redisStore)
	return &ThreadService{threadRepo: threadRepo, storage: storage, cache: cacheManager, logger: logger}
}

func (s *ThreadService) GetThreadByID(ctx context.Context, id string) (*ThreadDetail, error) {
	thread, err := s.threadRepo.GetThreadByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get thread: %w", err)
	}

	// Load tweets from IPFS
	tweets, err := s.loadTweetsFromIPFS(context.Background(), thread.CID)
	if err != nil {
		return nil, fmt.Errorf("load from ipfs %s: %w", thread.CID, err)
	}

	return &ThreadDetail{
		ID:             thread.ID,
		CID:            thread.CID,
		ContentPreview: thread.Summary,
		NumTweets:      thread.NumTweets,
		Tweets:         tweets,
		CreatedAt:      thread.CreatedAt,
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
