package model

import (
	"time"

	"gorm.io/gorm"
)

// PostImage represents an image in a post's content
type PostImage struct {
	ImageID string `json:"image_id"`
}

// Post represents a diary post record in the database
// The actual post content is stored in JSON files, this model only tracks the file location
type Post struct {
	ID       string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID   string `gorm:"not null;index;type:varchar(36)" json:"user_id"`
	FilePath string `gorm:"not null;size:500" json:"file_path"` // Path to the JSON file containing post content

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Foreign key relationships
	User UserProfile `gorm:"foreignKey:UserID;" json:"user"`
}

// TableName returns the table name for the Post model
func (Post) TableName() string {
	return "posts"
}
