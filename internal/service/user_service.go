package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"gorm.io/gorm"
)

// User Service Errors
var (
	// General user errors
	ErrUserNotFound = errors.New("user not found")
)

// UserProfileDetail represents a condensed user profile
type UserProfileDetail struct {
	model.UserProfile
}

// UserProfileSummary represents a condensed user profile
type UserProfileSummary struct {
	UserID    string  `json:"user_id"`
	DisplayID string  `json:"display_id"`
	Nickname  string  `json:"nickname"`
	Bio       *string `json:"bio"`
}

// PostSummary represents a brief post summary (placeholder)
type PostSummary struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	ContentPreview string    `json:"content_preview"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserRepoInterface defines the interface for user repo operations
type UserRepoInterface interface {
	GetUserByID(userID string) (*model.UserProfile, error)
	GetUserByDisplayID(displayID string) (*model.UserProfile, error)
	CreateUser(user *model.UserProfile) error
	UpdateUser(user *model.UserProfile) error
	DeleteUser(userID string) error
	SearchUsers(query string, limit, offset int) ([]model.UserProfile, int64, error)
}

// UserService provides business logic for user operations
type UserService struct {
	userRepo UserRepoInterface
}

// NewUserService creates a new user service
func NewUserService(userRepo UserRepoInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserProfile retrieves a user's profile
func (s *UserService) GetUserProfile(
	userID string,
) (*UserProfileDetail, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Build user profile
	profile := &UserProfileDetail{
		UserProfile: *user,
	}
	return profile, nil
}
