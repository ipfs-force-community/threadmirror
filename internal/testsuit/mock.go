package testsuit

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs/go-cid"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/datatypes"
)

// MockProcessedMarkRepo is a mock implementation for testing
type MockProcessedMarkRepo struct {
	processedMarks map[string]bool // key:type -> bool
}

func NewMockProcessedMarkRepo() *MockProcessedMarkRepo {
	return &MockProcessedMarkRepo{
		processedMarks: make(map[string]bool),
	}
}

func (m *MockProcessedMarkRepo) makeKey(key string, typ string) string {
	return key + ":" + typ
}

func (m *MockProcessedMarkRepo) IsProcessed(ctx context.Context, key string, typ string) (bool, error) {
	k := m.makeKey(key, typ)
	return m.processedMarks[k], nil
}

func (m *MockProcessedMarkRepo) MarkProcessed(ctx context.Context, key string, typ string) error {
	k := m.makeKey(key, typ)
	m.processedMarks[k] = true
	return nil
}

func (m *MockProcessedMarkRepo) BatchMarkProcessed(ctx context.Context, keys []string, typ string) error {
	for _, key := range keys {
		k := m.makeKey(key, typ)
		m.processedMarks[k] = true
	}
	return nil
}

// MockBotCookieRepo is a mock implementation for testing
type MockBotCookieRepo struct {
	cookies map[string][]byte // email:username -> JSON data
}

func NewMockBotCookieRepo() *MockBotCookieRepo {
	return &MockBotCookieRepo{
		cookies: make(map[string][]byte),
	}
}

func (m *MockBotCookieRepo) makeKey(email, username string) string {
	return email + ":" + username
}

func (m *MockBotCookieRepo) GetCookies(ctx context.Context, email, username string) (datatypes.JSON, error) {
	key := m.makeKey(email, username)
	cookies, exists := m.cookies[key]
	if !exists {
		return nil, errutil.ErrNotFound // Simulate no cookies found
	}
	return datatypes.JSON(cookies), nil
}

func (m *MockBotCookieRepo) SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error {
	// This would normally marshal the data in the real repo
	key := m.makeKey(email, username)
	m.cookies[key] = []byte(`[]`) // Store empty JSON for testing
	return nil
}

// GetLatestBotCookie returns an arbitrary cookie record (last added if any) to satisfy the interface.
func (m *MockBotCookieRepo) GetLatestBotCookie(ctx context.Context) (*model.BotCookie, error) {
	for key, data := range m.cookies {
		parts := strings.SplitN(key, ":", 2)
		email := parts[0]
		username := ""
		if len(parts) > 1 {
			username = parts[1]
		}
		return &model.BotCookie{
			Email:       email,
			Username:    username,
			CookiesData: datatypes.JSON(data),
		}, nil
	}
	return nil, errutil.ErrNotFound
}

// MockLLM is a mock implementation for testing
type MockLLM struct{}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Return a simple mock response
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "Mock AI summary for testing",
			},
		},
	}, nil
}

func (m *MockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "Mock AI summary for testing", nil
}

// MockIPFSStorage is a mock implementation for testing
type MockIPFSStorage struct{}

func (m *MockIPFSStorage) Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error) {
	// Return a fixed CID for testing
	c, _ := cid.Parse("bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u")
	return c, nil
}

func (m *MockIPFSStorage) Get(ctx context.Context, cid cid.Cid) (io.ReadCloser, error) {
	// Return mock JSON tweets data
	mockTweetsJSON := `[
		{
			"rest_id": "mock-tweet-1",
			"created_at": "2024-01-01T12:00:00Z",
			"text": "This is a mock tweet for testing",
			"author": {
				"rest_id": "mock-user-1",
				"name": "Mock User",
				"screen_name": "mockuser",
				"profile_image_url": "https://example.com/avatar.jpg"
			}
		}
	]`
	return io.NopCloser(strings.NewReader(mockTweetsJSON)), nil
}

// MockMentionRepo is a mock implementation for MentionRepoInterface
// Stores mentions in memory for testing
type MockMentionRepo struct {
	mentions map[string]*model.Mention
}

func NewMockMentionRepo() *MockMentionRepo {
	return &MockMentionRepo{
		mentions: make(map[string]*model.Mention),
	}
}

func (m *MockMentionRepo) GetMentionByID(ctx context.Context, id string) (*model.Mention, error) {
	mention, ok := m.mentions[id]
	if !ok {
		return nil, errutil.ErrNotFound
	}
	return mention, nil
}

func (m *MockMentionRepo) GetMentionByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Mention, error) {
	for _, mention := range m.mentions {
		if mention.UserID == userID && mention.ThreadID == threadID {
			return mention, nil
		}
	}
	return nil, errutil.ErrNotFound
}

func (m *MockMentionRepo) CreateMention(ctx context.Context, mention *model.Mention) error {
	m.mentions[mention.ID] = mention
	return nil
}

func (m *MockMentionRepo) GetMentions(ctx context.Context, userID string, limit, offset int) ([]model.Mention, int64, error) {
	var result []model.Mention
	for _, mention := range m.mentions {
		if userID == "" || mention.UserID == userID {
			result = append(result, *mention)
		}
	}
	total := int64(len(result))
	if offset > len(result) {
		offset = len(result)
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *MockMentionRepo) GetMentionsByUser(ctx context.Context, userID string, limit, offset int) ([]model.Mention, int64, error) {
	return m.GetMentions(ctx, userID, limit, offset)
}

// MockXScraper is a mock implementation for xscraper.Scraper interface
type MockXScraper struct {
	// Configuration for mock behavior
	ShouldReturnError bool
	MockTweets        []*xscraper.Tweet
	MockUsers         []*xscraper.User
}

func NewMockXScraper() *MockXScraper {
	// Create some default mock data
	mockUser := &xscraper.User{
		RestID:          "mock-user-123",
		ID:              "mock-user-123",
		Name:            "Mock User",
		ScreenName:      "mockuser",
		ProfileImageURL: "https://example.com/avatar.jpg",
		FollowersCount:  1000,
		FriendsCount:    500,
		StatusesCount:   250,
		Verified:        false,
		Description:     "This is a mock user for testing",
		CreatedAt:       time.Now(),
	}

	mockTweet := &xscraper.Tweet{
		RestID:    "mock-tweet-123",
		ID:        "mock-tweet-123",
		CreatedAt: time.Now(),
		Text:      "This is a mock tweet for testing purposes #testtweet",
		Author:    mockUser,
		Stats: xscraper.TweetStats{
			FavoriteCount: 10,
			RetweetCount:  5,
			ReplyCount:    2,
		},
	}

	return &MockXScraper{
		ShouldReturnError: false,
		MockTweets:        []*xscraper.Tweet{mockTweet},
		MockUsers:         []*xscraper.User{mockUser},
	}
}

func (m *MockXScraper) GetTweets(ctx context.Context, id string) (*xscraper.TweetsResult, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 404, Body: "Tweet not found"}
	}
	return &xscraper.TweetsResult{
		Tweets:     m.MockTweets,
		IsComplete: true, // Mock always returns complete results
	}, nil
}

func (m *MockXScraper) GetTweetDetail(ctx context.Context, id string) ([]*xscraper.Tweet, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 404, Body: "Tweet not found"}
	}
	return m.MockTweets, nil
}

func (m *MockXScraper) GetTweetResultByRestId(ctx context.Context, id string) (*xscraper.Tweet, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 404, Body: "Tweet not found"}
	}
	if len(m.MockTweets) > 0 {
		return m.MockTweets[0], nil
	}
	return nil, &xscraper.BadRequestError{StatusCode: 404, Body: "No mock tweets available"}
}

func (m *MockXScraper) SearchTweets(ctx context.Context, query string, maxTweets int) ([]*xscraper.Tweet, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 400, Body: "Search failed"}
	}

	// Return mock tweets up to maxTweets limit
	result := make([]*xscraper.Tweet, 0, maxTweets)
	for i, tweet := range m.MockTweets {
		if i >= maxTweets {
			break
		}
		result = append(result, tweet)
	}
	return result, nil
}

func (m *MockXScraper) CreateTweet(ctx context.Context, newTweet xscraper.NewTweet) (*xscraper.Tweet, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 400, Body: "Failed to create tweet"}
	}

	// Create a mock tweet based on the input
	mockTweet := &xscraper.Tweet{
		RestID:    "mock-created-tweet-123",
		ID:        "mock-created-tweet-123",
		CreatedAt: time.Now(),
		Text:      newTweet.Text,
		Author:    m.MockUsers[0], // Use first mock user
		Stats: xscraper.TweetStats{
			FavoriteCount: 0,
			RetweetCount:  0,
			ReplyCount:    0,
		},
	}

	return mockTweet, nil
}

func (m *MockXScraper) GetMentions(ctx context.Context, filter func(*xscraper.Tweet) bool) ([]*xscraper.Tweet, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 400, Body: "Failed to get mentions"}
	}

	var result []*xscraper.Tweet
	for _, tweet := range m.MockTweets {
		if filter == nil || filter(tweet) {
			result = append(result, tweet)
		}
	}
	return result, nil
}

func (m *MockXScraper) GetMentionsByScreenName(ctx context.Context, screenName string, filter func(*xscraper.Tweet) bool) ([]*xscraper.Tweet, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 400, Body: "Failed to get mentions"}
	}

	// Mock mentions for the specified screen name
	var result []*xscraper.Tweet
	for _, tweet := range m.MockTweets {
		// Simulate that this tweet mentions the screen name
		if strings.Contains(tweet.Text, "@"+screenName) || filter == nil || filter(tweet) {
			result = append(result, tweet)
		}
	}
	return result, nil
}

func (m *MockXScraper) UploadMedia(ctx context.Context, mediaReader io.Reader, mediaSize int) (*xscraper.MediaUploadResult, error) {
	if m.ShouldReturnError {
		return nil, &xscraper.BadRequestError{StatusCode: 400, Body: "Failed to upload media"}
	}

	return &xscraper.MediaUploadResult{
		MediaID: "mock-media-123456789",
		Size:    uint(mediaSize),
	}, nil
}

func (m *MockXScraper) Ready() bool {
	return !m.ShouldReturnError // Ready if not configured to return errors
}

func (m *MockXScraper) WaitForReady(ctx context.Context) error {
	if m.ShouldReturnError {
		return ctx.Err() // Return context error if configured to fail
	}
	return nil // Always ready in mock
}

// SetErrorMode configures the mock to return errors for testing error scenarios
func (m *MockXScraper) SetErrorMode(shouldError bool) {
	m.ShouldReturnError = shouldError
}

// AddMockTweet adds a custom tweet to the mock data
func (m *MockXScraper) AddMockTweet(tweet *xscraper.Tweet) {
	m.MockTweets = append(m.MockTweets, tweet)
}

// ClearMockData clears all mock data
func (m *MockXScraper) ClearMockData() {
	m.MockTweets = []*xscraper.Tweet{}
	m.MockUsers = []*xscraper.User{}
}
