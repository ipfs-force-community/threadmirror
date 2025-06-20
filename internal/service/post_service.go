package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Post Service Errors
var (
	ErrPostNotFound      = errors.New("post not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidContent    = errors.New("invalid content")
	ErrInvalidImageCount = errors.New("invalid image count")
	ErrInvalidImageID    = errors.New("invalid image ID")
)

// PostDetail represents a complete post with all details
type PostDetail struct {
	ID        datatypes.UUID     `json:"id"`
	Content   string             `json:"content"`
	User      UserProfileSummary `json:"user"`
	Images    []PostImageDetail  `json:"images"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// PostSummaryDetail represents a post summary for list views
type PostSummaryDetail struct {
	ID             datatypes.UUID     `json:"id"`
	ContentPreview string             `json:"content_preview"`
	User           UserProfileSummary `json:"user"`
	CreatedAt      time.Time          `json:"created_at"`
}

// PostImageDetail represents a post image
type PostImageDetail struct {
	ImageID string `json:"image_id"`
}

// PostImage represents a post image
type PostImage struct {
	ImageID string `json:"image_id"`
}

// CreatePostRequest represents a request to create a new post
type CreatePostRequest struct {
	Content  string   `json:"content"`
	ImageIDs []string `json:"image_ids"`
}

// UpdatePostRequest represents a request to update a post
type UpdatePostRequest struct {
	Content  *string  `json:"content"`
	ImageIDs []string `json:"image_ids"`
}

// PostRepoInterface defines the interface for post repo operations
type PostRepoInterface interface {
	// Post CRUD
	GetPostByID(id datatypes.UUID) (*model.Post, error)
	CreatePost(post *model.Post) error
	UpdatePost(post *model.Post) error
	DeletePost(id datatypes.UUID) error
	GetPosts(
		userID datatypes.UUID,
		limit, offset int,
	) ([]model.Post, int64, error)
	GetPostsByUser(userID datatypes.UUID, limit, offset int) ([]model.Post, int64, error)
}

// PostService provides business logic for post operations
type PostService struct {
	postRepo PostRepoInterface
	userRepo UserRepoInterface
}

// NewPostService creates a new post service
func NewPostService(
	postRepo PostRepoInterface,
	userRepo UserRepoInterface,
) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(
	userID datatypes.UUID,
	req *CreatePostRequest,
) (*PostDetail, error) {
	// Validate input
	if err := s.validatePostContent(req.Content); err != nil {
		return nil, err
	}

	if err := s.validateImageIDs(req.ImageIDs); err != nil {
		return nil, err
	}

	// Prepare images JSONB data
	var imagesJSON datatypes.JSON
	if len(req.ImageIDs) > 0 {
		images := make([]model.PostImage, len(req.ImageIDs))
		for i, id := range req.ImageIDs {
			images[i] = model.PostImage{
				ImageID: id,
			}
		}

		jsonData, err := json.Marshal(images)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal images: %w", err)
		}
		imagesJSON = datatypes.JSON(jsonData)
	} else {
		imagesJSON = datatypes.JSON([]byte("[]"))
	}

	// Create post
	post := &model.Post{
		Content: req.Content,
		UserID:  userID,
		Images:  imagesJSON,
	}

	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Return the created post details
	return s.GetPostByID(post.ID, userID)
}

// GetPostByID retrieves a post by ID
func (s *PostService) GetPostByID(
	postID datatypes.UUID,
	currentUserID datatypes.UUID,
) (*PostDetail, error) {
	post, err := s.postRepo.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return s.buildPostDetail(post, currentUserID, true)
}

// GetPosts retrieves posts based on feed type
func (s *PostService) GetPosts(
	userID datatypes.UUID,
	limit, offset int,
) ([]PostSummaryDetail, int64, error) {
	posts, _, err := s.postRepo.GetPosts(userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts: %w", err)
	}

	postSummaries := make([]PostSummaryDetail, len(posts))
	for i, post := range posts {
		summary, err := s.buildPostSummary(&post)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to build post summary: %w", err)
		}
		postSummaries[i] = *summary
	}

	return postSummaries, int64(len(posts)), nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(
	postID, userID datatypes.UUID,
	req *UpdatePostRequest,
) (*PostDetail, error) {
	// Get existing post
	post, err := s.postRepo.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Check ownership
	if post.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update fields if provided
	if req.Content != nil {
		if err := s.validatePostContent(*req.Content); err != nil {
			return nil, err
		}
		post.Content = *req.Content
	}

	// Update post images if provided
	if req.ImageIDs != nil {
		if err := s.validateImageIDs(req.ImageIDs); err != nil {
			return nil, err
		}

		// Update images JSONB data
		var imagesJSON datatypes.JSON
		if len(req.ImageIDs) > 0 {
			images := make([]model.PostImage, len(req.ImageIDs))
			for i, id := range req.ImageIDs {
				images[i] = model.PostImage{
					ImageID: id,
				}
			}

			jsonData, err := json.Marshal(images)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal images: %w", err)
			}
			imagesJSON = datatypes.JSON(jsonData)
		} else {
			imagesJSON = datatypes.JSON([]byte("[]"))
		}

		post.Images = imagesJSON
	}

	// Update the post
	if err := s.postRepo.UpdatePost(post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}
	// Return updated post
	return s.GetPostByID(postID, userID)
}

// DeletePost deletes a post
func (s *PostService) DeletePost(postID, userID datatypes.UUID) error {
	post, err := s.postRepo.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	if err := s.postRepo.DeletePost(postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

func (s *PostService) validatePostContent(content string) error {
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return ErrInvalidContent
	}
	if len(content) > 5000 {
		return ErrInvalidContent
	}
	return nil
}

func (s *PostService) validateImageIDs(ids []string) error {
	for _, id := range ids {
		if len(id) == 0 {
			return ErrInvalidImageID
		}
	}
	return nil
}

func (s *PostService) buildPostDetail(
	post *model.Post,
	currentUserID datatypes.UUID,
	includeComments bool,
) (*PostDetail, error) {
	// Parse images from JSONB
	var images []model.PostImage
	if len(post.Images) > 0 {
		if err := json.Unmarshal(post.Images, &images); err != nil {
			return nil, fmt.Errorf("failed to unmarshal post images: %w", err)
		}
	}

	imageDetails := make([]PostImageDetail, len(images))
	for i, img := range images {
		imageDetails[i] = PostImageDetail{
			ImageID: img.ImageID,
		}
	}

	return &PostDetail{
		ID:        post.ID,
		Content:   post.Content,
		User:      s.userToProfileSummary(&post.User),
		Images:    imageDetails,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

func (s *PostService) buildPostSummary(
	post *model.Post,
) (*PostSummaryDetail, error) {

	// Create content preview (first 200 characters)
	contentPreview := post.Content
	if len(contentPreview) > 200 {
		contentPreview = contentPreview[:200] + "..."
	}

	return &PostSummaryDetail{
		ID:             post.ID,
		ContentPreview: contentPreview,
		User:           s.userToProfileSummary(&post.User),
		CreatedAt:      post.CreatedAt,
	}, nil
}

func (s *PostService) userToProfileSummary(
	user *model.UserProfile,
) UserProfileSummary {
	return UserProfileSummary{
		UserID:    user.ID,
		DisplayID: user.DisplayID,
		Nickname:  user.Nickname,
		Bio:       user.Bio,
	}
}
