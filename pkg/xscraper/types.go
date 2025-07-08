package xscraper

import (
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
)

// User represents a simplified user structure for tweets
type User struct {
	ID              string    `json:"id"`
	RestID          string    `json:"rest_id"`
	Name            string    `json:"name"`
	ScreenName      string    `json:"screen_name"`
	ProfileImageURL string    `json:"profile_image_url"`
	Description     string    `json:"description"`
	FollowersCount  int       `json:"followers_count"`
	FriendsCount    int       `json:"friends_count"`
	StatusesCount   int       `json:"statuses_count"`
	CreatedAt       time.Time `json:"created_at"`
	Verified        bool      `json:"verified"`
	IsBlueVerified  bool      `json:"is_blue_verified"`
}

// TweetStats represents tweet engagement statistics
type TweetStats struct {
	ReplyCount    int `json:"reply_count"`
	RetweetCount  int `json:"retweet_count"`
	FavoriteCount int `json:"favorite_count"`
	QuoteCount    int `json:"quote_count"`
	BookmarkCount int `json:"bookmark_count"`
	ViewCount     int `json:"view_count,omitempty"`
}

// Tweet represents a simplified tweet structure
type Tweet struct {
	ID                string                             `json:"id"`
	RestID            string                             `json:"rest_id"`
	Text              string                             `json:"text"`
	CreatedAt         time.Time                          `json:"created_at"`
	Author            *User                              `json:"author,omitempty"`
	Entities          generated.Entities                 `json:"entities"`
	Stats             TweetStats                         `json:"stats"`
	IsRetweet         bool                               `json:"is_retweet"`
	IsReply           bool                               `json:"is_reply"`
	IsQuoteStatus     bool                               `json:"is_quote_status"`
	ConversationID    string                             `json:"conversation_id"`
	InReplyToStatusID string                             `json:"in_reply_to_status_id,omitempty"`
	InReplyToUserID   string                             `json:"in_reply_to_user_id,omitempty"`
	QuotedTweet       *Tweet                             `json:"quoted_tweet,omitempty"`
	HasBirdwatchNotes bool                               `json:"has_birdwatch_notes"`
	Lang              string                             `json:"lang"`
	Source            string                             `json:"source,omitempty"`
	PossiblySensitive bool                               `json:"possibly_sensitive"`
	IsTranslatable    bool                               `json:"is_translatable"`
	Views             int                                `json:"views,omitempty"`
	IsNoteTweet       bool                               `json:"is_note_tweet"`
	RichText          *generated.NoteTweetResultRichText `json:"richtext,omitempty"`
}

// convertTimelineToTweets converts a generated.Timeline to our Tweet struct
func convertTimelineToTweets(timeline *generated.Timeline) ([]*Tweet, error) {
	if timeline == nil {
		return nil, fmt.Errorf("timeline is nil")
	}

	if len(timeline.Instructions) == 0 {
		return nil, fmt.Errorf("no instructions found in timeline")
	}

	var tweets []*Tweet
	convertAndAppendTweet := func(timelineTweet *generated.TimelineTweet) error {
		if timelineTweet.TweetResults.Result == nil {
			return nil
		}
		// Try to get Tweet from TweetUnion
		tweetData, err := timelineTweet.TweetResults.Result.AsTweet()
		if err != nil {
			return nil
		}

		tweet, err := convertGeneratedTweetToTweet(&tweetData)
		if err != nil {
			return fmt.Errorf("failed to convert tweet data: %w", err)
		}
		tweets = append(tweets, tweet)

		return nil
	}

	// Iterate through instructions to find TimelineAddEntries
	for _, instruction := range timeline.Instructions {
		addEntries, err := instruction.AsTimelineAddEntries()
		if err != nil {
			continue // Skip if not TimelineAddEntries
		}
		// Look for the main tweet entry
		for _, entry := range addEntries.Entries {
			// Try to get TimelineTimelineModule
			if timelineModule, err := entry.Content.AsTimelineTimelineModule(); err == nil && timelineModule.Items != nil {
				for _, item := range *timelineModule.Items {
					timelineTweet, err := item.Item.ItemContent.AsTimelineTweet()
					if err != nil {
						continue // Skip if not TimelineTimelineItem
					}
					if err := convertAndAppendTweet(&timelineTweet); err != nil {
						return nil, err
					}
				}
			}

			// Try to get TimelineTimelineItem
			if timelineItem, err := entry.Content.AsTimelineTimelineItem(); err == nil {
				// Try to get TimelineTweet from ItemContent
				timelineTweet, err := timelineItem.ItemContent.AsTimelineTweet()
				if err != nil {
					continue // Skip if not TimelineTweet
				}
				if err := convertAndAppendTweet(&timelineTweet); err != nil {
					return nil, err
				}
			}
		}
	}

	return tweets, nil
}

// convertGeneratedTweetToTweet converts generated.Tweet to our Tweet struct
func convertGeneratedTweetToTweet(genTweet *generated.Tweet) (*Tweet, error) {
	if genTweet == nil {
		return nil, fmt.Errorf("generated tweet is nil")
	}

	tweet := &Tweet{
		RestID: genTweet.RestId,
		ID:     genTweet.RestId,
	}

	// Extract basic tweet information from Legacy
	if genTweet.Legacy != nil {
		legacy := genTweet.Legacy
		tweet.Text = legacy.FullText
		tweet.ConversationID = legacy.ConversationIdStr
		tweet.IsQuoteStatus = legacy.IsQuoteStatus
		tweet.Lang = legacy.Lang
		tweet.PossiblySensitive = legacy.PossiblySensitive != nil && *legacy.PossiblySensitive

		// Parse created at
		if createdAt, err := time.Parse(time.RubyDate, string(legacy.CreatedAt)); err == nil {
			tweet.CreatedAt = createdAt
		}

		// Set reply information
		if legacy.InReplyToStatusIdStr != nil {
			tweet.InReplyToStatusID = *legacy.InReplyToStatusIdStr
			tweet.IsReply = true
		}
		if legacy.InReplyToUserIdStr != nil {
			tweet.InReplyToUserID = *legacy.InReplyToUserIdStr
		}

		// Set engagement stats
		tweet.Stats = TweetStats{
			ReplyCount:    legacy.ReplyCount,
			RetweetCount:  legacy.RetweetCount,
			FavoriteCount: legacy.FavoriteCount,
			QuoteCount:    legacy.QuoteCount,
			BookmarkCount: legacy.BookmarkCount,
		}

		tweet.Entities = legacy.Entities

	}

	if genTweet.NoteTweet != nil {
		// Long-form Note Tweet overrides legacy text and entities
		note := genTweet.NoteTweet.NoteTweetResults.Result

		// Replace text and entities with the content from the note tweet
		tweet.Text = note.Text
		tweet.Entities = note.EntitySet
		tweet.IsNoteTweet = true
		tweet.RichText = note.Richtext
	}

	// Extract user information from Core
	if genTweet.Core != nil && genTweet.Core.UserResults.Result != nil {
		userResult := genTweet.Core.UserResults.Result
		user, err := userResult.AsUser()
		if err == nil {
			tweet.Author = convertGeneratedUserToUser(&user)
		}
	}

	// Set additional flags
	if genTweet.HasBirdwatchNotes != nil {
		tweet.HasBirdwatchNotes = *genTweet.HasBirdwatchNotes
	}
	if genTweet.IsTranslatable != nil {
		tweet.IsTranslatable = *genTweet.IsTranslatable
	}
	if genTweet.Source != nil {
		tweet.Source = *genTweet.Source
	}

	// Extract view count if available
	if genTweet.Views != nil && genTweet.Views.Count != nil {
		if count, err := fmt.Sscanf(*genTweet.Views.Count, "%d", &tweet.Views); count == 1 && err == nil {
			// View count successfully parsed and assigned to tweet.Views
			tweet.Stats.ViewCount = tweet.Views // Also update the Stats field for backward compatibility
		}
	}

	// Handle quoted tweet if present
	if genTweet.QuotedStatusResult != nil && genTweet.QuotedStatusResult.Result != nil {
		quotedTweet, err := genTweet.QuotedStatusResult.Result.AsTweet()
		if err == nil {
			tweet.QuotedTweet, _ = convertGeneratedTweetToTweet(&quotedTweet)
		}
	}

	return tweet, nil
}

// convertGeneratedUserToUser converts generated.User to our User struct
func convertGeneratedUserToUser(genUser *generated.User) *User {
	if genUser == nil {
		return nil
	}

	legacy := &genUser.Legacy
	user := &User{
		RestID:         genUser.RestId,
		ID:             genUser.Id,
		Name:           legacy.Name,
		ScreenName:     legacy.ScreenName,
		Description:    legacy.Description,
		FollowersCount: legacy.FollowersCount,
		FriendsCount:   legacy.FriendsCount,
		StatusesCount:  legacy.StatusesCount,
		Verified:       legacy.Verified,
		IsBlueVerified: genUser.IsBlueVerified,
	}

	// Parse created at
	if createdAt, err := time.Parse(time.RubyDate, string(legacy.CreatedAt)); err == nil {
		user.CreatedAt = createdAt
	}

	// Set profile image URL
	user.ProfileImageURL = legacy.ProfileImageUrlHttps

	return user
}
