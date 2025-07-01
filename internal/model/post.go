package model

import (
	"time"
)

type Post struct {
	ID       string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID   string `gorm:"not null;index;type:varchar(36)" json:"user_id"`
	ThreadID string `gorm:"not null;index;type:varchar(36)" json:"thread_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Post model
func (Post) TableName() string {
	return "posts"
}
