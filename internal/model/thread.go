package model

import (
	"time"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

// TweetWithTranslation wraps xscraper.Tweet with translation information
type TweetWithTranslation struct {
	*xscraper.Tweet `json:",inline"`
	Translations    map[string]string `json:"translations,omitempty"` // language code -> translated text
}

// ThreadStatus represents the current status of thread scraping
type ThreadStatus string

const (
	ThreadStatusPending   ThreadStatus = "pending"
	ThreadStatusScraping  ThreadStatus = "scraping"
	ThreadStatusCompleted ThreadStatus = "completed"
	ThreadStatusFailed    ThreadStatus = "failed"
)

type Thread struct {
	ID      string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Summary string `gorm:"not null;size:500" json:"summary"` // Summary of the post
	CID     string `gorm:"not null;size:64" json:"cid"`      // CID of the post content

	NumTweets int `gorm:"not null;default:0" json:"numTweets"` // Number of tweets in the thread

	// Thread status tracking
	Status ThreadStatus `gorm:"not null;default:'pending';type:varchar(20)" json:"status"`

	// Retry tracking for cron jobs (internal use only)
	RetryCount int `gorm:"not null;default:0" json:"-"` // Number of retry attempts

	// Optimistic locking
	Version int `gorm:"not null;default:1" json:"version"`

	// Thread author information (centralized here instead of duplicating in mentions)
	AuthorID              string `gorm:"type:varchar(36);index" json:"author_id"`
	AuthorName            string `gorm:"size:100" json:"author_name"`              // Display name
	AuthorScreenName      string `gorm:"size:50" json:"author_screen_name"`        // Screen name (without @)
	AuthorProfileImageURL string `gorm:"size:500" json:"author_profile_image_url"` // Profile image URL

	// Has one translation relationship (GORM will auto-detect ThreadID foreign key)
	Translation *Translation `json:"translation,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Thread) TableName() string {
	return "threads"
}
