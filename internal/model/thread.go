package model

import "time"

type Thread struct {
	ID      string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Summary string `gorm:"not null;size:500" json:"summary"` // Summary of the post
	CID     string `gorm:"not null;size:64" json:"cid"`      // CID of the post content

	// Author information from the original tweet
	AuthorID              string `gorm:"not null;index;type:varchar(36)" json:"author_id"`
	AuthorName            string `gorm:"size:100" json:"author_name"`              // Display name
	AuthorScreenName      string `gorm:"size:50" json:"author_screen_name"`        // Screen name (without @)
	AuthorProfileImageURL string `gorm:"size:500" json:"author_profile_image_url"` // Profile image URL

	NumTweets int `gorm:"not null;default:0" json:"numTweets"` // Number of tweets in the thread

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Thread) TableName() string {
	return "threads"
}
