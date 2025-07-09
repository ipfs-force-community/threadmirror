package testsuit

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisModule "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerTestSuite provides a test suite using real database and Redis containers
type ContainerTestSuite struct {
	DB             *sql.DB
	RedisClient    *redis.Client
	MentionRepo    *sqlrepo.MentionRepo
	MentionService *service.MentionService
	ThreadRepo     *sqlrepo.ThreadRepo

	// Container references for cleanup
	pgContainer    testcontainers.Container
	redisContainer testcontainers.Container
}

// SetupContainerTestSuite creates a new test suite with testcontainers
func SetupContainerTestSuite(t *testing.T) *ContainerTestSuite {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start PostgreSQL container")

	// Get PostgreSQL connection string
	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get PostgreSQL connection string")

	// Start Redis container
	redisContainer, err := redisModule.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start Redis container")

	// Get Redis connection details
	redisHost, err := redisContainer.Host(ctx)
	require.NoError(t, err, "Failed to get Redis host")

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err, "Failed to get Redis port")

	// Create database connection
	db, err := sql.New("postgres", pgConnStr, slog.Default())
	require.NoError(t, err, "Failed to connect to PostgreSQL")

	// Migrate the schema
	err = db.AutoMigrate(model.AllModels()...)
	require.NoError(t, err, "Failed to migrate database schema")

	// Create Redis client
	redisConfig := &redis.RedisConfig{
		Addr:     redisHost + ":" + redisPort.Port(),
		Password: "",
		DB:       0,
	}
	redisClient := redis.NewClient(redisConfig)

	// Test Redis connection
	err = redisClient.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis")

	// Create repositories
	mentionRepo := sqlrepo.NewMentionRepo(db)
	threadRepo := sqlrepo.NewThreadRepo(db)

	// Create mock dependencies
	mockLLM := &MockLLM{}
	mockIPFS := &MockIPFSStorage{}

	// Create services
	mentionService := service.NewMentionService(
		mentionRepo,
		llm.Model(mockLLM),
		ipfs.Storage(mockIPFS),
		threadRepo,
		db,
	)

	return &ContainerTestSuite{
		DB:             db,
		RedisClient:    redisClient,
		MentionRepo:    mentionRepo,
		MentionService: mentionService,
		ThreadRepo:     threadRepo,
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
	}
}

// TearDown cleans up all containers and connections
func (s *ContainerTestSuite) TearDown(t *testing.T) {
	ctx := context.Background()

	// Close database connection
	if s.DB != nil {
		err := s.DB.Close()
		if err != nil {
			t.Logf("Warning: Failed to close database connection: %v", err)
		}
	}

	// Close Redis connection
	if s.RedisClient != nil {
		err := s.RedisClient.Close()
		if err != nil {
			t.Logf("Warning: Failed to close Redis connection: %v", err)
		}
	}

	// Terminate containers
	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(ctx)
		if err != nil {
			t.Logf("Warning: Failed to terminate PostgreSQL container: %v", err)
		}
	}

	if s.redisContainer != nil {
		err := s.redisContainer.Terminate(ctx)
		if err != nil {
			t.Logf("Warning: Failed to terminate Redis container: %v", err)
		}
	}
}

// ResetDatabase cleans all tables for test isolation
func (s *ContainerTestSuite) ResetDatabase(t *testing.T) {
	ctx := context.Background()

	// Start a transaction to clean up all data
	tx := s.DB.Begin()
	defer tx.Rollback()

	// Delete data from all tables in reverse dependency order
	tables := []string{
		"mentions",
		"threads",
		"processed_marks",
		"bot_cookies",
	}

	for _, table := range tables {
		err := tx.WithContext(ctx).Exec("DELETE FROM " + table).Error
		require.NoError(t, err, "Failed to clean table %s", table)
	}

	err := tx.Commit().Error
	require.NoError(t, err, "Failed to commit database cleanup")
}

// ResetRedis clears all Redis data for test isolation
func (s *ContainerTestSuite) ResetRedis(t *testing.T) {
	ctx := context.Background()

	err := s.RedisClient.FlushAll(ctx).Err()
	require.NoError(t, err, "Failed to flush Redis data")
}

// Reset resets both database and Redis for complete test isolation
func (s *ContainerTestSuite) Reset(t *testing.T) {
	s.ResetDatabase(t)
	s.ResetRedis(t)
}
