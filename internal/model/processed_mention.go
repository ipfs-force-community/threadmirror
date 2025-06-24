package model

import (
	"time"
)

// ProcessedMention tracks processed mentions to avoid duplicate responses
type ProcessedMention struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    string    `gorm:"not null;size:255;index:idx_user_tweet,priority:1" json:"user_id"`
	TweetID   string    `gorm:"not null;size:255;index:idx_user_tweet,priority:2" json:"tweet_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for ProcessedMention
func (ProcessedMention) TableName() string {
	return "processed_mentions"
}
