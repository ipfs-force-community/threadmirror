package sqlrepo

import (
	"context"
	"errors"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"gorm.io/gorm"
)

// PostRepo implements PostRepoInterface
type PostRepo struct{}

// NewPostRepo creates a new post repo
func NewPostRepo() *PostRepo {
	return &PostRepo{}
}

// Post CRUD operations

// GetPostByID retrieves a post by ID
func (r *PostRepo) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	db := sql.MustDBFromContext(ctx)
	var post model.Post
	err := db.Where("id = ?", id).First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepo) GetPostByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Post, error) {
	db := sql.MustDBFromContext(ctx)
	var post model.Post
	err := db.Where("user_id = ? AND thread_id = ?", userID, threadID).First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

// CreatePost creates a new post
func (r *PostRepo) CreatePost(ctx context.Context, post *model.Post) error {
	db := sql.MustDBFromContext(ctx)
	return db.Create(post).Error
}

// GetPosts retrieves posts based on feed type with optional filtering
func (r *PostRepo) GetPosts(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]model.Post, int64, error) {
	db := sql.MustDBFromContext(ctx)
	var posts []model.Post
	var total int64

	query := db.Model(&model.Post{})

	// Filter by user if provided
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("posts.created_at DESC").Limit(limit).Offset(offset).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsByUser retrieves posts created by a specific user with pagination
func (r *PostRepo) GetPostsByUser(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]model.Post, int64, error) {
	db := sql.MustDBFromContext(ctx)
	var posts []model.Post
	var total int64

	query := db.Model(&model.Post{}).
		Where("user_id = ?", userID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}
