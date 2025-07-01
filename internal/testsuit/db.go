package testsuit

import (
	"log/slog"
	"testing"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/stretchr/testify/require"
)

func SetupTestDB(t *testing.T) *sql.DB {
	db, err := sql.New("sqlite", ":memory:", slog.Default())
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(model.AllModels()...)
	require.NoError(t, err)

	return db
}

// DBTestSuite provides a test suite based on real database instead of mocks
type DBTestSuite struct {
	DB          *sql.DB
	PostRepo    *sqlrepo.PostRepo
	PostService *service.PostService
}

// SetupDBTestSuite creates a new database test suite with all dependencies
func SetupDBTestSuite(t *testing.T) *DBTestSuite {
	// Setup test database
	db := SetupTestDB(t)

	// Create repositories
	postRepo := sqlrepo.NewPostRepo()
	threadRepo := sqlrepo.NewThreadRepo()

	// Create mock dependencies
	mockLLM := &MockLLM{}
	mockIPFS := &MockIPFSStorage{}

	// Create services
	postService := service.NewPostService(postRepo, llm.Model(mockLLM), ipfs.Storage(mockIPFS), threadRepo, db)

	return &DBTestSuite{
		DB:          db,
		PostRepo:    postRepo,
		PostService: postService,
	}
}
