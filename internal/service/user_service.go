package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"gorm.io/datatypes"
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
	UserID    datatypes.UUID `json:"user_id"`
	DisplayID string         `json:"display_id"`
	Nickname  string         `json:"nickname"`
	Bio       *string        `json:"bio"`
}

// UpdateUserProfileRequest represents a user profile update request
type UpdateUserProfileRequest struct {
	Nickname *string `json:"nickname"`
	Bio      *string `json:"bio"`
	Email    *string `json:"email"`
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
	GetUserByID(userID datatypes.UUID) (*model.UserProfile, error)
	GetUserByDisplayID(displayID string) (*model.UserProfile, error)
	CreateUser(user *model.UserProfile) error
	UpdateUser(user *model.UserProfile) error
	DeleteUser(userID datatypes.UUID) error
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
	userID datatypes.UUID,
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

// UpdateUserProfile updates a user's profile information
func (s *UserService) UpdateUserProfile(
	userID datatypes.UUID,
	req *UpdateUserProfileRequest,
) (*UserProfileDetail, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.Email != nil {
		user.Email = req.Email
	}

	// Save updated user
	if err := s.userRepo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Return updated profile
	return s.GetUserProfile(userID)
}

// userToProfileSummary converts a UserProfile model to UserProfileSummary with follow status
func (s *UserService) userToProfileSummary(
	user *model.UserProfile,
	currentUserID datatypes.UUID,
) UserProfileSummary {
	summary := UserProfileSummary{
		UserID:    user.ID,
		DisplayID: user.DisplayID,
		Nickname:  user.Nickname,
	}

	// Truncate bio for summary
	if user.Bio != nil {
		bio := *user.Bio
		if len(bio) > 50 {
			truncated := bio[:47] + "..."
			summary.Bio = &truncated
		} else {
			summary.Bio = user.Bio
		}
	}

	return summary
}
