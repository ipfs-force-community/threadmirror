package testsuit

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/ipfs-force-community/threadmirror/internal/api/v1"
	v1middleware "github.com/ipfs-force-community/threadmirror/internal/api/v1/middleware"
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

// SetupTestServer sets up a test server with the given database
func SetupTestServer(t *testing.T, db *sql.DB) *gin.Engine {
	userRepo := sqlrepo.NewUserRepo(db)
	postRepo := sqlrepo.NewPostRepo(db)
	userSvc := service.NewUserService(userRepo)
	postSvc := service.NewPostService(postRepo, userRepo)
	logger := slog.New(slog.NewTextHandler(nil, nil))
	server := v1.NewV1Handler(userSvc, postSvc, createTestSupabaseConfig(), logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id for tests
	router.Use(func(c *gin.Context) {
		// Set different user IDs based on the request path for different tests
		path := c.Request.URL.Path
		if strings.Contains(path, "/users/user1/") || strings.Contains(path, "/users/user2/") {
			SetTestAuthInfo(
				c,
				datatypes.NewUUIDv4(),
			) // For follow/unfollow tests, set current user as user1
		} else {
			SetTestAuthInfo(c, datatypes.NewUUIDv4()) // For profile tests
		}
		c.Next()
	})

	// Add error handling middleware (inline to avoid import cycle)
	router.Use(v1middleware.ErrorHandler())

	v1.RegisterHandlers(router, server)

	return router
}
