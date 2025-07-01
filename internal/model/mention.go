package model

import (
	"time"
)

type Mention struct {
	ID       string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID   string `gorm:"not null;uniqueIndex:idx_user_id_thread_id;type:varchar(36)" json:"user_id"`
	ThreadID string `gorm:"not null;uniqueIndex:idx_user_id_thread_id;type:varchar(36)" json:"thread_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Mention model
func (Mention) TableName() string {
	return "mentions"
}
