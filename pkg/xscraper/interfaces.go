package xscraper

import (
	"context"
	"io"
)

// Scraper defines the interface for Twitter scraping operations
// This interface allows for easy mocking in tests
type Scraper interface {
	// Tweet operations
	GetTweets(ctx context.Context, id string) ([]*Tweet, error)
	GetTweetDetail(ctx context.Context, id string) ([]*Tweet, error)
	GetTweetResultByRestId(ctx context.Context, id string) (*Tweet, error)
	SearchTweets(ctx context.Context, query string, maxTweets int) ([]*Tweet, error)
	CreateTweet(ctx context.Context, newTweet NewTweet) (*Tweet, error)

	// Mention operations
	GetMentions(ctx context.Context, filter func(*Tweet) bool) ([]*Tweet, error)
	GetMentionsByScreenName(ctx context.Context, screenName string, filter func(*Tweet) bool) ([]*Tweet, error)

	// Media operations
	UploadMedia(ctx context.Context, mediaReader io.Reader, mediaSize int) (*MediaUploadResult, error)

	// Utility operations
	Ready() bool
	WaitForReady(ctx context.Context) error
}

// Ensure XScraper implements the Scraper interface
var _ Scraper = (*XScraper)(nil)
