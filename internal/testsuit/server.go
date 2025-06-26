package testsuit

import (
	"log/slog"
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
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
)

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

// SetupTestServer sets up a test server with the given database
func SetupTestServer(t *testing.T, db *sql.DB) *gin.Engine {
	postRepo := sqlrepo.NewPostRepo(db)

	// Create mock dependencies
	mockLLM := &MockLLM{}
	mockIPFS := &MockIPFSStorage{}

	postSvc := service.NewPostService(postRepo, llm.Model(mockLLM), ipfs.Storage(mockIPFS))
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

	server := v1.NewV1Handler(postSvc, logger, twitterBot)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add error handling middleware (inline to avoid import cycle)
	router.Use(v1middleware.ErrorHandler())

	v1.RegisterHandlers(router, server)

	return router
}
