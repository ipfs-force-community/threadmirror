package model

import "time"

type Thread struct {
	ID      string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Summary string `gorm:"not null;size:500" json:"summary"` // Summary of the post
	CID     string `gorm:"not null;size:64" json:"cid"`      // CID of the post content

	NumTweets int `gorm:"not null;default:0" json:"numTweets"` // Number of tweets in the thread

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Thread) TableName() string {
	return "threads"
}
