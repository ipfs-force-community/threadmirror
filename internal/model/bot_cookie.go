package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BotCookie stores Twitter bot session cookies in database
type BotCookie struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Email       string         `gorm:"not null;size:255;uniqueIndex:idx_email_username" json:"email"`
	Username    string         `gorm:"not null;size:255;uniqueIndex:idx_email_username" json:"username"`
	CookiesData datatypes.JSON `gorm:"type:jsonb" json:"cookies_data"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName returns the table name for BotCookie
func (BotCookie) TableName() string {
	return "bot_cookies"
}
