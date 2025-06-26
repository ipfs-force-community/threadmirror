package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"gorm.io/gorm"
)

// Post Service Errors
var (
	ErrPostNotFound = errors.New("post not found")
	ErrInvalidPath  = errors.New("invalid file path")
)

// PostDetail represents a complete post with all details
type PostDetail struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Images    []PostImageDetail `json:"images"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// PostSummaryDetail represents a post summary for list views
type PostSummaryDetail struct {
	ID             string    `json:"id"`
	ContentPreview string    `json:"content_preview"`
	CreatedAt      time.Time `json:"created_at"`
}

// PostImageDetail represents a post image
type PostImageDetail struct {
	ImageID string `json:"image_id"`
}

// CreatePostRequest represents a request to create a new post
type CreatePostRequest struct {
	FilePath string `json:"file_path"`
}

// PostRepoInterface defines the interface for post repo operations
type PostRepoInterface interface {
	// Post CRUD
	GetPostByID(id string) (*model.Post, error)
	CreatePost(post *model.Post) error
	GetPosts(
		userID string,
		limit, offset int,
	) ([]model.Post, int64, error)
	GetPostsByUser(userID string, limit, offset int) ([]model.Post, int64, error)
}

// PostService provides business logic for post operations
type PostService struct {
	postRepo PostRepoInterface
}

// NewPostService creates a new post service
func NewPostService(
	postRepo PostRepoInterface,
) *PostService {
	return &PostService{
		postRepo: postRepo,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(
	userID string,
	req *CreatePostRequest,
) (*PostDetail, error) {
	// Validate input
	if err := s.validateFilePath(req.FilePath); err != nil {
		return nil, err
	}

	// Create post record
	post := &model.Post{
		UserID:   userID,
		FilePath: req.FilePath,
	}

	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Return the created post details
	return s.GetPostByID(post.ID)
}

// GetPostByID retrieves a post by ID
func (s *PostService) GetPostByID(postID string) (*PostDetail, error) {
	post, err := s.postRepo.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return s.buildPostDetail(post)
}

// GetPosts retrieves posts based on feed type
func (s *PostService) GetPosts(
	userID string,
	limit, offset int,
) ([]PostSummaryDetail, int64, error) {
	posts, total, err := s.postRepo.GetPosts(userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts: %w", err)
	}

	postSummaries := make([]PostSummaryDetail, 0, len(posts))
	for _, post := range posts {
		postSummaries = append(postSummaries, *s.buildPostSummary(&post))
	}

	return postSummaries, total, nil
}

// validateFilePath validates file path
func (s *PostService) validateFilePath(filePath string) error {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return ErrInvalidPath
	}
	// Add more validation logic as needed (e.g., check file extension, path format)
	return nil
}

// buildPostDetail builds a PostDetail from a model.Post
func (s *PostService) buildPostDetail(
	post *model.Post,
) (*PostDetail, error) {
	filePath := post.FilePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &PostDetail{
		ID:        post.ID,
		Content:   string(content),
		Images:    []PostImageDetail{},
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

// buildPostSummary builds a PostSummaryDetail from a model.Post
func (s *PostService) buildPostSummary(
	post *model.Post,
) *PostSummaryDetail {
	return &PostSummaryDetail{
		ID:             post.ID,
		ContentPreview: "", // Content preview should be loaded from file_path when needed
		CreatedAt:      post.CreatedAt,
	}
}
