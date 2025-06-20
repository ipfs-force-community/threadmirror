package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PostImage represents an image in a post's JSONB array
type PostImage struct {
	ImageID string `json:"image_id"`
}

// Post represents a diary post in the database
type Post struct {
	ID      datatypes.UUID `gorm:"primaryKey" json:"id"`
	Content string         `gorm:"not null;size:1000"          json:"content"`
	UserID  datatypes.UUID `gorm:"not null;index"              json:"user_id"`
	Images  datatypes.JSON `gorm:"type:jsonb"                  json:"images"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"          gorm:"index"`

	// Foreign key relationships
	User UserProfile `gorm:"foreignKey:UserID;"            json:"user"`
}

// TableName returns the table name for the Post model
func (Post) TableName() string {
	return "posts"
}

// BeforeCreate hook to generate ID if not provided
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID.IsEmpty() {
		p.ID = datatypes.NewUUIDv4()
	}
	return nil
}
