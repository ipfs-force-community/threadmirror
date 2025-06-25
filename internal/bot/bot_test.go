package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

// MockProcessedMentionRepo is a mock implementation for testing
type MockProcessedMentionRepo struct {
	processed map[string]bool // userID:tweetID -> bool
}

func NewMockProcessedMentionRepo() *MockProcessedMentionRepo {
	return &MockProcessedMentionRepo{
		processed: make(map[string]bool),
	}
}

func (m *MockProcessedMentionRepo) makeKey(userID string, tweetID string) string {
	return userID + ":" + tweetID
}

func (m *MockProcessedMentionRepo) IsProcessed(ctx context.Context, userID string, tweetID string) (bool, error) {
	key := m.makeKey(userID, tweetID)
	return m.processed[key], nil
}

func (m *MockProcessedMentionRepo) MarkProcessed(ctx context.Context, userID string, tweetID string) error {
	key := m.makeKey(userID, tweetID)
	m.processed[key] = true
	return nil
}

func (m *MockProcessedMentionRepo) BatchMarkProcessed(ctx context.Context, userID string, tweetIDs []string) error {
	for _, tweetID := range tweetIDs {
		key := m.makeKey(userID, tweetID)
		m.processed[key] = true
	}
	return nil
}

// MockBotCookieRepo is a mock implementation for testing
type MockBotCookieRepo struct {
	cookies []*http.Cookie // Directly store http.Cookie slice
}

func NewMockBotCookieRepo() *MockBotCookieRepo {
	return &MockBotCookieRepo{
		cookies: nil,
	}
}

func (m *MockBotCookieRepo) GetCookies(ctx context.Context, email, username string) (datatypes.JSON, error) {
	if m.cookies == nil {
		return nil, nil // Simulate no cookies found
	}
	// Marshal cookies to JSON for return
	jsonData, err := json.Marshal(m.cookies)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(jsonData), nil
}

func (m *MockBotCookieRepo) SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error {
	// Convert cookiesData to []*http.Cookie
	if cookies, ok := cookiesData.([]*http.Cookie); ok {
		m.cookies = cookies
		return nil
	}
	return fmt.Errorf("invalid cookie data type")
}

func createTestBot(_ *testing.T) *TwitterBot {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create mock services
	mockProcessedMentionRepo := NewMockProcessedMentionRepo()
	processedMentionService := service.NewProcessedMentionService(mockProcessedMentionRepo)

	mockBotCookieRepo := NewMockBotCookieRepo()
	botCookieService := service.NewBotCookieService(mockBotCookieRepo)

	return NewTwitterBot(
		"testbot",          // username
		"testpass",         // password
		"test@example.com", // email
		5*time.Minute,      // checkInterval
		10,                 // maxMentionsCheck
		processedMentionService,
		botCookieService,
		logger,
	)
}

func TestNewTwitterBot(t *testing.T) {
	bot := createTestBot(t)

	assert.NotNil(t, bot)
	assert.NotNil(t, bot.scraper)
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
	assert.Equal(t, "database", stats["storage_type"])
}
