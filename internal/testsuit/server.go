package testsuit

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/ipfs-force-community/threadmirror/internal/api/v1"
	v1middleware "github.com/ipfs-force-community/threadmirror/internal/api/v1/middleware"
	"github.com/ipfs-force-community/threadmirror/internal/bot"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"gorm.io/datatypes"
)

// createTestSupabaseConfig creates a test supabase config for testing
func createTestSupabaseConfig() *config.SupabaseConfig {
	return &config.SupabaseConfig{
		ProjectReference: "test-project-ref",
		ApiAnnoKey:       "test-api-key",
		BucketNames: config.SupabaseBucketNames{
			PostImages: "post-images",
		},
	}
}

// createTestBotConfig creates a test bot config for testing
func createTestBotConfig() *config.BotConfig {
	return &config.BotConfig{
		Username:         "testbot",
		Password:         "testpass",
		Email:            "test@example.com",
		CheckInterval:    5 * time.Minute,
		MaxMentionsCheck: 10,
	}
}

// MockProcessedMentionRepo is a mock implementation for testing
type MockProcessedMentionRepo struct {
	processedMentions map[string]bool // userID:tweetID -> bool
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
		return nil, nil // Simulate no cookies found
	}
	return datatypes.JSON(cookies), nil
}

func (m *MockBotCookieRepo) SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error {
	// This would normally marshal the data in the real repo
	key := m.makeKey(email, username)
	m.cookies[key] = []byte(`[]`) // Store empty JSON for testing
	return nil
}

// SetupTestServer sets up a test server with the given database
func SetupTestServer(t *testing.T, db *sql.DB) *gin.Engine {
	userRepo := sqlrepo.NewUserRepo(db)
	postRepo := sqlrepo.NewPostRepo(db)
	userSvc := service.NewUserService(userRepo)
	postSvc := service.NewPostService(postRepo, userRepo)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	// Mock processed mention repo and service
	mockProcessedMentionRepo := &MockProcessedMentionRepo{
		processedMentions: make(map[string]bool),
	}
	processedMentionService := service.NewProcessedMentionService(mockProcessedMentionRepo)

	// Mock bot cookie repo and service
	mockBotCookieRepo := &MockBotCookieRepo{
		cookies: make(map[string][]byte),
	}
	botCookieService := service.NewBotCookieService(mockBotCookieRepo)

	// Create test bot
	testBotConfig := createTestBotConfig()
	twitterBot := bot.NewTwitterBot(
		testBotConfig.Username,
		testBotConfig.Password,
		testBotConfig.Email,
		testBotConfig.CheckInterval,
		testBotConfig.MaxMentionsCheck,
		processedMentionService,
		botCookieService,
		logger,
	)

	server := v1.NewV1Handler(userSvc, postSvc, createTestSupabaseConfig(), logger, twitterBot)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id for tests
	router.Use(func(c *gin.Context) {
		// Set different user IDs based on the request path for different tests
		path := c.Request.URL.Path
		if strings.Contains(path, "/users/user1/") || strings.Contains(path, "/users/user2/") {
			SetTestAuthInfo(
				c,
				datatypes.NewUUIDv4().String(),
			) // For follow/unfollow tests, set current user as user1
		} else {
			SetTestAuthInfo(c, datatypes.NewUUIDv4().String()) // For profile tests
		}
		c.Next()
	})

	// Add error handling middleware (inline to avoid import cycle)
	router.Use(v1middleware.ErrorHandler())

	v1.RegisterHandlers(router, server)

	return router
}
