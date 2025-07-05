package model

import (
	"time"
)

// ProcessedMark tracks processed business actions to avoid duplicate responses
type ProcessedMark struct {
	Key       string    `gorm:"size:36;primaryKey" json:"key"`
	Type      string    `gorm:"size:32;primaryKey" json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for ProcessedMark
func (ProcessedMark) TableName() string {
	return "processed_marks"
}
