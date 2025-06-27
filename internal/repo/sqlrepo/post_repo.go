package sqlrepo

import (
	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

// PostRepo implements PostRepoInterface
type PostRepo struct {
	db *sql.DB
}

// NewPostRepo creates a new post repo
func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{db: db}
}

// Post CRUD operations

// GetPostByID retrieves a post by ID with preloaded relationships
func (r *PostRepo) GetPostByID(id string) (*model.Post, error) {
	var post model.Post
	err := r.db.Where("id = ?", id).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// CreatePost creates a new post
func (r *PostRepo) CreatePost(post *model.Post) error {
	return r.db.Create(post).Error
}

// GetPosts retrieves posts based on feed type with optional filtering
func (r *PostRepo) GetPosts(
	userID string,
	limit, offset int,
) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	query := r.db.Model(&model.Post{})

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
	userID string,
	limit, offset int,
) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	query := r.db.Model(&model.Post{}).
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
