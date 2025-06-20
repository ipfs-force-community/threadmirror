package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserProfile represents a user profile in the database
type UserProfile struct {
	ID        datatypes.UUID `gorm:"primaryKey"                          json:"id"`
	DisplayID string         `gorm:"uniqueIndex;not null;size:50"        json:"display_id"`
	Nickname  string         `gorm:"not null;default:'nameless';size:20" json:"nickname"`
	Bio       *string        `gorm:"size:100"                            json:"bio"`
	Email     *string        `gorm:"size:255"                            json:"email"`

	// Cached social statistics
	PostsCount int64 `gorm:"not null;default:0" json:"posts_count"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"          gorm:"index"`
}

// TableName returns the table name for the UserProfile model
func (UserProfile) TableName() string {
	return "user_profiles"
}
