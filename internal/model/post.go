package model

import (
	"time"
)

// Post represents a diary post record in the database
// The actual post content is stored in JSON files, this model only tracks the file location
type Post struct {
	ID      string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID  string `gorm:"not null;index;type:varchar(36)" json:"user_id"`
	Summary string `gorm:"not null;size:500" json:"summary"` // Summary of the post
	CID     string `gorm:"not null;size:64" json:"cid"`      // CID of the post content

	// Author information from the original tweet
	AuthorID              string `gorm:"not null;index;type:varchar(36)" json:"author_id"`
	AuthorName            string `gorm:"size:100" json:"author_name"`              // Display name
	AuthorScreenName      string `gorm:"size:50" json:"author_screen_name"`        // Screen name (without @)
	AuthorProfileImageURL string `gorm:"size:500" json:"author_profile_image_url"` // Profile image URL

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Post model
func (Post) TableName() string {
	return "posts"
}
