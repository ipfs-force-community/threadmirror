package model

import (
	"time"
)

type Mention struct {
	ID              string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID          string    `gorm:"not null;uniqueIndex:idx_user_id_thread_id;type:varchar(36)" json:"user_id"`
	ThreadID        string    `gorm:"not null;uniqueIndex:idx_user_id_thread_id;type:varchar(36)" json:"thread_id"`
	MentionCreateAt time.Time `json:"mention_create_at"`

	// GORM association - Thread relationship
	Thread Thread `gorm:"foreignKey:ThreadID;references:ID" json:"thread,omitempty"`

	// Author information is now stored in Thread table to avoid duplication
	// ThreadAuthor can be accessed via Thread.Author* fields

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Mention model
func (Mention) TableName() string {
	return "mentions"
}
