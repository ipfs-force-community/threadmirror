package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/datatypes"
)

// mockJobQueueClient implements jobq.JobQueueClient for testing
type mockJobQueueClient struct{}

func (m *mockJobQueueClient) Enqueue(ctx context.Context, job *jobq.Job) (string, error) {
	return "mock-job-id", nil
}

// MockProcessedMentionRepo is a mock implementation for testing
// userID:tweetID -> bool
type MockProcessedMentionRepo struct {
	processedMentions map[string]bool
}

func NewMockProcessedMentionRepo() *MockProcessedMentionRepo {
	return &MockProcessedMentionRepo{
		processedMentions: make(map[string]bool),
	}
}

func (m *MockProcessedMentionRepo) makeKey(userID string, tweetID string) string {
	return userID + ":" + tweetID
}

func (m *MockProcessedMentionRepo) IsProcessed(ctx context.Context, userID string, tweetID string) (bool, error) {
	key := m.makeKey(userID, tweetID)
	return m.processedMentions[key], nil
}

func (m *MockProcessedMentionRepo) MarkProcessed(ctx context.Context, userID string, tweetID string) error {
	key := m.makeKey(userID, tweetID)
	m.processedMentions[key] = true
	return nil
}

func (m *MockProcessedMentionRepo) BatchMarkProcessed(ctx context.Context, userID string, tweetIDs []string) error {
	for _, tweetID := range tweetIDs {
		key := m.makeKey(userID, tweetID)
		m.processedMentions[key] = true
	}
	return nil
}

// MockBotCookieRepo is a mock implementation for testing
// email:username -> JSON data
type MockBotCookieRepo struct {
	cookies map[string][]byte
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
		return nil, nil // Simulate no cookies found
	}
	return datatypes.JSON(cookies), nil
}

func (m *MockBotCookieRepo) SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error {
	key := m.makeKey(email, username)
	data, err := json.Marshal(cookiesData)
	if err != nil {
		return err
	}
	m.cookies[key] = data
	return nil
}

// MockPostRepo is a mock implementation for PostRepoInterface
// Stores posts in memory for testing
type MockPostRepo struct {
	posts map[string]*model.Post
}

func NewMockPostRepo() *MockPostRepo {
	return &MockPostRepo{
		posts: make(map[string]*model.Post),
	}
}

func (m *MockPostRepo) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

func (m *MockPostRepo) CreatePost(ctx context.Context, post *model.Post) error {
	m.posts[post.ID] = post
	return nil
}

func (m *MockPostRepo) GetPosts(ctx context.Context, userID string, limit, offset int) ([]model.Post, int64, error) {
	var result []model.Post
	for _, post := range m.posts {
		if userID == "" || post.UserID == userID {
			result = append(result, *post)
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

func (m *MockPostRepo) GetPostsByUser(ctx context.Context, userID string, limit, offset int) ([]model.Post, int64, error) {
	return m.GetPosts(ctx, userID, limit, offset)
}

// MockLLM is a mock implementation for testing
// Implements llm.Model (alias of llms.Model)
type MockLLM struct{}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: "Mock AI summary for testing"}},
	}, nil
}

func (m *MockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "Mock AI summary for testing", nil
}

// MockIPFSStorage is a mock implementation for testing
// Implements ipfs.Storage
type MockIPFSStorage struct{}

func (m *MockIPFSStorage) Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error) {
	c, _ := cid.Parse("bafkreiabc123")
	return c, nil
}

func (m *MockIPFSStorage) Get(ctx context.Context, c cid.Cid) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock content")), nil
}

func createTestBot(_ *testing.T) *TwitterBot {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create mock services
	mockProcessedMentionRepo := NewMockProcessedMentionRepo()
	processedMentionService := service.NewProcessedMentionService(mockProcessedMentionRepo)

	mockBotCookieRepo := NewMockBotCookieRepo()
	botCookieService := service.NewBotCookieService(mockBotCookieRepo)

	// Use new mocks for PostService
	mockPostRepo := NewMockPostRepo()
	mockLLM := &MockLLM{}
	mockIPFS := &MockIPFSStorage{}

	// Pass nil for threadRepo and db (not used in these tests)
	mockPostService := service.NewPostService(mockPostRepo, mockLLM, mockIPFS, nil, nil)

	// Mock JobQueueClient
	jobQueueClient := &mockJobQueueClient{}

	return NewTwitterBot(
		"testbot",          // username
		"test@example.com", // email
		nil,                // scraper
		5*time.Minute,      // checkInterval
		10,                 // maxMentionsCheck
		processedMentionService,
		botCookieService,
		mockPostService,
		jobQueueClient,
		logger,
	)
}

func TestNewTwitterBot(t *testing.T) {
	bot := createTestBot(t)

	assert.NotNil(t, bot)
	// bot.scraper 允许为 nil，因为测试未注入 mock scraper
	// assert.NotNil(t, bot.scraper)
	assert.NotNil(t, bot.botCookieService)
	assert.NotNil(t, bot.processedMentionService)
	assert.NotNil(t, bot.logger)
}

func TestGenerateResponse(t *testing.T) {
	// Skip this test as generateResponse method has been removed
	t.Skip("generateResponse method no longer exists")
}

func TestProcessedMentionService(t *testing.T) {
	mockRepo := NewMockProcessedMentionRepo()
	processedMentionService := service.NewProcessedMentionService(mockRepo)
	ctx := context.Background()
	userID := "VXNlcjoxNDAzODgxMTMwODAyMjI1MTUy"

	// Test initially not processed
	isProcessed, err := processedMentionService.IsProcessed(ctx, userID, "tweet1")
	assert.NoError(t, err)
	assert.False(t, isProcessed)

	// Mark as processed
	err = processedMentionService.MarkProcessed(ctx, userID, "tweet1")
	assert.NoError(t, err)

	isProcessed, err = processedMentionService.IsProcessed(ctx, userID, "tweet1")
	assert.NoError(t, err)
	assert.True(t, isProcessed)

	// Test different tweet
	isProcessed, err = processedMentionService.IsProcessed(ctx, userID, "tweet2")
	assert.NoError(t, err)
	assert.False(t, isProcessed)

	// Test different user, same tweet
	userID2 := "VXNlcjoyMjkzODgxMTMwODAyMjI1MTUy"
	isProcessed, err = processedMentionService.IsProcessed(ctx, userID2, "tweet1")
	assert.NoError(t, err)
	assert.False(t, isProcessed) // Should be false for different user
}

func TestBotCookieService(t *testing.T) {
	mockRepo := NewMockBotCookieRepo()
	botCookieService := service.NewBotCookieService(mockRepo)
	ctx := context.Background()
	testEmail := "test@example.com"
	testUsername := "testbot"

	// Test initially no cookies
	cookies, err := botCookieService.LoadCookies(ctx, testEmail, testUsername)
	assert.NoError(t, err)
	assert.Nil(t, cookies)

	// Save some test cookies
	testCookies := []*http.Cookie{
		{
			Name:  "test_cookie",
			Value: "test_value",
		},
	}
	err = botCookieService.SaveCookies(ctx, testEmail, testUsername, testCookies)
	assert.NoError(t, err)

	// Load cookies - should now return the saved cookies
	cookies, err = botCookieService.LoadCookies(ctx, testEmail, testUsername)
	assert.NoError(t, err)
	assert.NotNil(t, cookies)
	assert.Len(t, cookies, 1)
	assert.Equal(t, "test_cookie", cookies[0].Name)
	assert.Equal(t, "test_value", cookies[0].Value)
}

func TestGetStats(t *testing.T) {
	bot := createTestBot(t)

	stats := bot.GetStats()

	assert.Equal(t, true, stats["enabled"])
	assert.Equal(t, "testbot", stats["username"])
	assert.Equal(t, "5m0s", stats["check_interval"])
	assert.Equal(t, true, stats["randomized"])
	assert.Equal(t, "database", stats["storage_type"])
}

func TestRandomizedInterval(t *testing.T) {
	bot := createTestBot(t)
	baseInterval := bot.checkInterval // 5 minutes

	// Test multiple calculations to ensure randomization
	intervals := make([]time.Duration, 20)
	for i := 0; i < 20; i++ {
		intervals[i] = bot.randomizedInterval()
	}

	// Calculate expected range (±30% of base interval)
	jitterRange := time.Duration(float64(baseInterval) * 0.3)
	minExpected := baseInterval - jitterRange
	maxExpected := baseInterval + jitterRange

	// Ensure minimum of 30 seconds
	if minExpected < 30*time.Second {
		minExpected = 30 * time.Second
	}

	for _, interval := range intervals {
		// Should be within expected bounds
		assert.GreaterOrEqual(t, interval, minExpected,
			"Interval should be >= min expected (%v)", minExpected)
		assert.LessOrEqual(t, interval, maxExpected,
			"Interval should be <= max expected (%v)", maxExpected)
	}

	// Check that we get variation (not all the same)
	allSame := true
	for i := 1; i < len(intervals); i++ {
		if intervals[i] != intervals[0] {
			allSame = false
			break
		}
	}
	assert.False(t, allSame, "Should generate different intervals, got all same: %v", intervals[0])
}
