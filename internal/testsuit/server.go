package testsuit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/ipfs-force-community/threadmirror/internal/api/v1"
	"github.com/ipfs-force-community/threadmirror/internal/bot"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n"
	"github.com/ipfs-force-community/threadmirror/pkg/job"
	"github.com/ipfs-force-community/threadmirror/pkg/log"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// NewTestServer creates a new fx.App for testing purposes.
// It allows overriding specific modules or providing additional dependencies.
func NewTestServer(ctx context.Context, opts ...fx.Option) (*fx.App, error) {
	// Default test configuration
	serverCfg := &config.ServerConfig{
		Addr: "localhost:8080", // Use Addr instead of Port
	}
	dbCfg := &config.DatabaseConfig{
		DSN: "sqlite://:memory:?_foreign_keys=on", // Use in-memory SQLite for tests
	}
	botCfg := &config.BotConfig{
		Username:             "testbot",
		Password:             "testpass",
		Email:                "test@example.com",
		CheckInterval:        60 * time.Second, // Use CheckInterval
		MaxMentionsCheck:     10,
	}
	authCfg := &config.AuthConfig{
		Audience: "test-audience",
		Domain:   "test-domain",
	}
	logCfg := &config.LogConfig{
		Level: "debug",
	}

	// Mock Asynq client for testing
	testJobQueueClient := &MockJobQueueClient{}

	// Base modules for the test server
	baseModules := []fx.Option{
		fx.Logger(fxevent.NopLogger),
		fx.Supply(serverCfg, dbCfg, botCfg, authCfg, logCfg), // Supply individual config structs
		log.Module,
		sql.Module,
		service.Module,
		v1.Module,
		auth.Module,
		i18n.Module,
		fx.Provide(func() job.JobQueueClient { return testJobQueueClient }), // Provide mock Asynq client
		fx.Provide(func() job.JobQueueServer { return &MockJobQueueServer{} }), // Provide mock Asynq server
		fx.Provide(func(lc fx.Lifecycle, logger *slog.Logger) bot.XScraper {
			// Mock XScraper for tests
			return &MockXScraper{}
		}),
		fx.Provide(func(processedMentionService *service.ProcessedMentionService, botCookieService *service.BotCookieService, logger *slog.Logger, scraper bot.XScraper) *bot.TwitterBot {
			return bot.NewTwitterBot(
				botCfg.Username,
				botCfg.Password,
				botCfg.Email,
				botCfg.CheckInterval,
				botCfg.MaxMentionsCheck,
				processedMentionService,
				botCookieService,
				testJobQueueClient, // Use the mock client
				scraper, // Inject scraper
				logger,
			)
		}),
		fx.Invoke(func(*http.Server) {}), // Prevent server from starting
	}

	// Combine base modules with provided options
	allOpts := append(baseModules, opts...)

	app := fx.New(allOpts...)

	// Start the app with a context that can be cancelled
	if err := app.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start test server: %w", err)
	}

	return app, nil
}

// MockJobQueueClient for testing purposes
type MockJobQueueClient struct {
	mock.Mock
}

func (m *MockJobQueueClient) Enqueue(job *job.Job, opts ...job.EnqueueOption) (*asynq.TaskInfo, error) {
	args := m.Called(job, opts)
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

func (m *MockJobQueueClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockJobQueueServer for testing purposes
type MockJobQueueServer struct {
	mock.Mock
}

func (m *MockJobQueueServer) RegisterHandler(taskType string, handler job.JobHandler) {
	m.Called(taskType, handler)
}

func (m *MockJobQueueServer) Run() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockJobQueueServer) Shutdown() {
	m.Called()
}

type MockXScraper struct {
	mock.Mock
}

func (m *MockXScraper) GetMentions(ctx context.Context) ([]*xscraper.Tweet, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*xscraper.Tweet), args.Error(1)
}

func (m *MockXScraper) Login(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}