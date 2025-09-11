package testsuit

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisModule "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerConfig holds configuration for test containers
type ContainerConfig struct {
	PostgresImage  string
	RedisImage     string
	DBName         string
	DBUser         string
	DBPassword     string
	StartupTimeout time.Duration
}

// DefaultContainerConfig returns default configuration for test containers
func DefaultContainerConfig() *ContainerConfig {
	return &ContainerConfig{
		PostgresImage:  "postgres:15-alpine",
		RedisImage:     "redis:7-alpine",
		DBName:         "testdb",
		DBUser:         "testuser",
		DBPassword:     "testpass",
		StartupTimeout: 60 * time.Second,
	}
}

// SetupPostgresContainer creates and starts a PostgreSQL test container
func SetupPostgresContainer(t *testing.T, config *ContainerConfig) (testcontainers.Container, *sql.DB) {
	if config == nil {
		config = DefaultContainerConfig()
	}

	ctx := context.Background()

	// Start PostgreSQL container
	container, err := postgres.Run(ctx,
		config.PostgresImage,
		postgres.WithDatabase(config.DBName),
		postgres.WithUsername(config.DBUser),
		postgres.WithPassword(config.DBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(config.StartupTimeout),
		),
	)
	require.NoError(t, err, "Failed to start PostgreSQL container")

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get PostgreSQL connection string")

	// Create database connection
	db, err := sql.New("postgres", connStr, slog.Default())
	require.NoError(t, err, "Failed to connect to PostgreSQL")

	// SQLC architecture uses declarative schema files
	// Migration is handled by Supabase CLI, not needed in tests

	t.Logf("PostgreSQL container started successfully: %s", connStr)

	return container, db
}

// SetupRedisContainer creates and starts a Redis test container
func SetupRedisContainer(t *testing.T, config *ContainerConfig) (testcontainers.Container, *redis.Client) {
	if config == nil {
		config = DefaultContainerConfig()
	}

	ctx := context.Background()

	// Start Redis container
	container, err := redisModule.Run(ctx,
		config.RedisImage,
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(config.StartupTimeout),
		),
	)
	require.NoError(t, err, "Failed to start Redis container")

	// Get connection details
	host, err := container.Host(ctx)
	require.NoError(t, err, "Failed to get Redis host")

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err, "Failed to get Redis port")

	// Create Redis client
	redisConfig := &redis.RedisConfig{
		Addr:     host + ":" + port.Port(),
		Password: "",
		DB:       0,
	}
	client := redis.NewClient(redisConfig)

	// Test connection
	err = client.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis")

	t.Logf("Redis container started successfully: %s:%s", host, port.Port())

	return container, client
}

// CleanupContainer safely terminates a container
func CleanupContainer(t *testing.T, container testcontainers.Container) {
	if container == nil {
		return
	}

	ctx := context.Background()
	err := container.Terminate(ctx)
	if err != nil {
		t.Logf("Warning: Failed to terminate container: %v", err)
	}
}

// CleanupDatabase safely closes a database connection
func CleanupDatabase(t *testing.T, db *sql.DB) {
	if db == nil {
		return
	}

	err := db.Close()
	if err != nil {
		t.Logf("Warning: Failed to close database connection: %v", err)
	}
}

// CleanupRedis safely closes a Redis connection
func CleanupRedis(t *testing.T, client *redis.Client) {
	if client == nil {
		return
	}

	err := client.Close()
	if err != nil {
		t.Logf("Warning: Failed to close Redis connection: %v", err)
	}
}

// SkipIfContainerUnavailable skips the test if Docker is not available
func SkipIfContainerUnavailable(t *testing.T) {
	if os.Getenv("SKIP_CONTAINER_TESTS") != "" {
		t.Skip("Container tests are disabled")
	}

	// Check if running in CI without Docker
	if isCI() && !hasDocker() {
		t.Skip("Docker is not available in CI environment")
	}
}

// isCI checks if running in a CI environment
func isCI() bool {
	ci := os.Getenv("CI")
	if ci == "" {
		return false
	}

	if result, err := strconv.ParseBool(ci); err == nil {
		return result
	}

	// Common CI environment variables
	ciVars := []string{
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_URL",
		"CIRCLE_CI",
		"TRAVIS",
	}

	for _, envVar := range ciVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// hasDocker checks if Docker is available
func hasDocker() bool {
	// This is a simple check - in a real implementation you might want to
	// actually try to connect to Docker daemon
	return os.Getenv("DOCKER_HOST") != "" ||
		fileExists("/var/run/docker.sock") ||
		fileExists("\\\\.\\pipe\\docker_engine") // Windows
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// WaitForHealthy waits for a container to become healthy
func WaitForHealthy(t *testing.T, container testcontainers.Container, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Container did not become healthy within %v", timeout)
		case <-ticker.C:
			state, err := container.State(ctx)
			if err != nil {
				t.Logf("Failed to get container state: %v", err)
				continue
			}

			if state.Running {
				return
			}
		}
	}
}

// GetContainerLogs retrieves logs from a container for debugging
func GetContainerLogs(t *testing.T, container testcontainers.Container) string {
	ctx := context.Background()

	logs, err := container.Logs(ctx)
	if err != nil {
		t.Logf("Failed to get container logs: %v", err)
		return ""
	}
	defer func() {
		if closeErr := logs.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close container logs: %v", closeErr)
		}
	}()

	buf := make([]byte, 1024*10) // 10KB buffer
	n, err := logs.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Logf("Failed to read container logs: %v", err)
		return ""
	}

	return string(buf[:n])
}

// CreateTestEnvironment creates a complete test environment with both PostgreSQL and Redis
func CreateTestEnvironment(t *testing.T) (*ContainerTestSuite, func()) {
	SkipIfContainerUnavailable(t)

	suite := SetupContainerTestSuite(t)

	cleanup := func() {
		suite.TearDown(t)
	}

	// Register cleanup with t.Cleanup for automatic execution
	t.Cleanup(cleanup)

	return suite, cleanup
}
