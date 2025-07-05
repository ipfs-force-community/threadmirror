package model

import (
	"time"
)

type Mention struct {
	ID              string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID          string    `gorm:"not null;uniqueIndex:idx_user_id_thread_id;type:varchar(36)" json:"user_id"`
	ThreadID        string    `gorm:"not null;uniqueIndex:idx_user_id_thread_id;type:varchar(36)" json:"thread_id"`
	MentionCreateAt time.Time `json:"mention_create_at"`

	// Author information from the thread
	ThreadAuthorID              string `gorm:"not null;index;type:varchar(36)" json:"thread_author_id"`
	ThreadAuthorName            string `gorm:"size:100" json:"thread_author_name"`              // Display name
	ThreadAuthorScreenName      string `gorm:"size:50" json:"thread_author_screen_name"`        // Screen name (without @)
	ThreadAuthorProfileImageURL string `gorm:"size:500" json:"thread_author_profile_image_url"` // Profile image URL

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Mention model
func (Mention) TableName() string {
	return "mentions"
}
