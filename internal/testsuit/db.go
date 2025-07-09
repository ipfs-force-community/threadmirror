package testsuit

import (
	"context"
	"log"
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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisModule "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Global containers for ginkgo tests
var (
	ginkgoSuite *ContainerTestSuite
)

// SetupTestDB (original version) 使用testcontainers创建真实的PostgreSQL数据库用于测试
func SetupTestDB(t *testing.T) *sql.DB {
	SkipIfContainerUnavailable(t)

	suite := SetupContainerTestSuite(t)
	// 注意：这里不调用defer suite.TearDown(t)
	// 因为调用者可能需要在测试后续使用数据库
	// 调用者应该负责清理
	t.Cleanup(func() {
		suite.TearDown(t)
	})

	return suite.DB
}

// SetupGinkgoTestDB creates a PostgreSQL testcontainer for ginkgo tests
func SetupGinkgoTestDB() *sql.DB {
	if ginkgoSuite == nil {
		ginkgoSuite = setupGinkgoContainerSuite()
	}
	return ginkgoSuite.DB
}

// SetupTestRedis creates a Redis testcontainer for ginkgo tests
func SetupTestRedis() *redis.Client {
	if ginkgoSuite == nil {
		ginkgoSuite = setupGinkgoContainerSuite()
	}
	return ginkgoSuite.RedisClient
}

// CleanupTestContainers cleans up all test containers (for ginkgo AfterSuite)
func CleanupTestContainers() {
	if ginkgoSuite != nil {
		ctx := context.Background()

		// Close database connection
		if ginkgoSuite.DB != nil {
			if err := ginkgoSuite.DB.Close(); err != nil {
				log.Printf("Warning: Failed to close database connection: %v", err)
			}
		}

		// Close Redis connection
		if ginkgoSuite.RedisClient != nil {
			if err := ginkgoSuite.RedisClient.Close(); err != nil {
				log.Printf("Warning: Failed to close Redis connection: %v", err)
			}
		}

		// Terminate containers
		if ginkgoSuite.pgContainer != nil {
			if err := ginkgoSuite.pgContainer.Terminate(ctx); err != nil {
				log.Printf("Warning: Failed to terminate PostgreSQL container: %v", err)
			}
		}

		if ginkgoSuite.redisContainer != nil {
			if err := ginkgoSuite.redisContainer.Terminate(ctx); err != nil {
				log.Printf("Warning: Failed to terminate Redis container: %v", err)
			}
		}

		ginkgoSuite = nil
	}
}

// setupGinkgoContainerSuite creates a new test suite for ginkgo tests
func setupGinkgoContainerSuite() *ContainerTestSuite {
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
	if err != nil {
		panic("Failed to start PostgreSQL container: " + err.Error())
	}

	// Get PostgreSQL connection string
	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("Failed to get PostgreSQL connection string: " + err.Error())
	}

	// Start Redis container
	redisContainer, err := redisModule.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic("Failed to start Redis container: " + err.Error())
	}

	// Get Redis connection details
	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		panic("Failed to get Redis host: " + err.Error())
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		panic("Failed to get Redis port: " + err.Error())
	}

	// Create database connection
	db, err := sql.New("postgres", pgConnStr, slog.Default())
	if err != nil {
		panic("Failed to connect to PostgreSQL: " + err.Error())
	}

	// Migrate the schema
	err = db.AutoMigrate(model.AllModels()...)
	if err != nil {
		panic("Failed to migrate database schema: " + err.Error())
	}

	// Create Redis client
	redisConfig := &redis.RedisConfig{
		Addr:     redisHost + ":" + redisPort.Port(),
		Password: "",
		DB:       0,
	}
	redisClient := redis.NewClient(redisConfig)

	// Test Redis connection
	err = redisClient.Ping(ctx).Err()
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

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

// DBTestSuite provides a test suite based on real database instead of mocks
// 已弃用：使用 ContainerTestSuite 替代，它提供更完整的testcontainers环境
type DBTestSuite struct {
	DB             *sql.DB
	MentionRepo    *sqlrepo.MentionRepo
	MentionService *service.MentionService
}

// SetupDBTestSuite creates a new database test suite with all dependencies
// 已弃用：使用 SetupContainerTestSuite 替代，它提供更完整的testcontainers环境
func SetupDBTestSuite(t *testing.T) *DBTestSuite {
	SkipIfContainerUnavailable(t)

	// Setup testcontainers environment
	suite := SetupContainerTestSuite(t)
	t.Cleanup(func() {
		suite.TearDown(t)
	})

	// Create repositories
	mentionRepo := sqlrepo.NewMentionRepo(suite.DB)
	threadRepo := sqlrepo.NewThreadRepo(suite.DB)

	// Create mock dependencies
	mockLLM := &MockLLM{}
	mockIPFS := &MockIPFSStorage{}

	// Create services
	mentionService := service.NewMentionService(mentionRepo, llm.Model(mockLLM), ipfs.Storage(mockIPFS), threadRepo, suite.DB)

	return &DBTestSuite{
		DB:             suite.DB,
		MentionRepo:    mentionRepo,
		MentionService: mentionService,
	}
}

// ResetGinkgoDatabase cleans all tables for test isolation in ginkgo tests
func ResetGinkgoDatabase() {
	if ginkgoSuite == nil {
		return
	}

	ctx := context.Background()

	// Start a transaction to clean up all data
	tx := ginkgoSuite.DB.Begin()
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
		if err != nil {
			panic("Failed to clean table " + table + ": " + err.Error())
		}
	}

	err := tx.Commit().Error
	if err != nil {
		panic("Failed to commit database cleanup: " + err.Error())
	}
}
