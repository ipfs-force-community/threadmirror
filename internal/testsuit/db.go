package testsuit

import (
	"log/slog"
	"testing"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
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
	UserRepo    *sqlrepo.UserRepo
	PostRepo    *sqlrepo.PostRepo
	UserService *service.UserService
	PostService *service.PostService
}

// SetupDBTestSuite creates a new database test suite with all dependencies
func SetupDBTestSuite(t *testing.T) *DBTestSuite {
	// Setup test database
	db := SetupTestDB(t)

	// Create repositories
	userRepo := sqlrepo.NewUserRepo(db)
	postRepo := sqlrepo.NewPostRepo(db)

	// Create services
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userRepo)

	return &DBTestSuite{
		DB:          db,
		UserRepo:    userRepo,
		PostRepo:    postRepo,
		UserService: userService,
		PostService: postService,
	}
}
