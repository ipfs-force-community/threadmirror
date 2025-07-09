package xscraper

import (
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
)

// isHashtagEqual compares two hashtags for equality based on text field
func isHashtagEqual(a, b generated.Hashtag) bool {
	textA, okA := a["text"]
	textB, okB := b["text"]
	return okA && okB && textA == textB
}

// isSymbolEqual compares two symbols for equality based on text field
func isSymbolEqual(a, b generated.Symbol) bool {
	textA, okA := a["text"]
	textB, okB := b["text"]
	return okA && okB && textA == textB
}

// isUserMentionEqual compares two user mentions for equality based on screen_name field
func isUserMentionEqual(a, b generated.UserMention) bool {
	screenNameA, okA := a["screen_name"]
	screenNameB, okB := b["screen_name"]
	return okA && okB && screenNameA == screenNameB
}

// isMediaEqual compares two media items for equality based on unique identifiers
func isMediaEqual(a, b generated.Media) bool {
	return a.IdStr == b.IdStr && a.MediaKey == b.MediaKey
}

// isUrlEqual compares two URLs for equality based on URL
func isUrlEqual(a, b generated.Url) bool {
	return a.Url == b.Url
}

// isTimestampEqual compares two timestamps for equality
func isTimestampEqual(a, b generated.Timestamp) bool {
	return a.Seconds == b.Seconds && a.Text == b.Text
}

// containsHashtag checks if a hashtag exists in a slice
func containsHashtag(slice []generated.Hashtag, item generated.Hashtag) bool {
	for _, h := range slice {
		if isHashtagEqual(h, item) {
			return true
		}
	}
	return false
}

// containsSymbol checks if a symbol exists in a slice
func containsSymbol(slice []generated.Symbol, item generated.Symbol) bool {
	for _, s := range slice {
		if isSymbolEqual(s, item) {
			return true
		}
	}
	return false
}

// containsUserMention checks if a user mention exists in a slice
func containsUserMention(slice []generated.UserMention, item generated.UserMention) bool {
	for _, u := range slice {
		if isUserMentionEqual(u, item) {
			return true
		}
	}
	return false
}

// containsMedia checks if a media item exists in a slice
func containsMedia(slice []generated.Media, item generated.Media) bool {
	for _, m := range slice {
		if isMediaEqual(m, item) {
			return true
		}
	}
	return false
}

// containsUrl checks if a URL exists in a slice
func containsUrl(slice []generated.Url, item generated.Url) bool {
	for _, u := range slice {
		if isUrlEqual(u, item) {
			return true
		}
	}
	return false
}

// containsTimestamp checks if a timestamp exists in a slice
func containsTimestamp(slice []generated.Timestamp, item generated.Timestamp) bool {
	for _, t := range slice {
		if isTimestampEqual(t, item) {
			return true
		}
	}
	return false
}

// MergeEntities merges two Entities, preserving all entities from both sources without duplication
func MergeEntities(original, updated generated.Entities) generated.Entities {
	merged := generated.Entities{
		Hashtags:     make([]generated.Hashtag, 0),
		Symbols:      make([]generated.Symbol, 0),
		UserMentions: make([]generated.UserMention, 0),
		Urls:         make([]generated.Url, 0),
	}

	// Merge hashtags - add all from original, then add new ones from updated
	merged.Hashtags = append(merged.Hashtags, original.Hashtags...)
	for _, h := range updated.Hashtags {
		if !containsHashtag(merged.Hashtags, h) {
			merged.Hashtags = append(merged.Hashtags, h)
		}
	}

	// Merge symbols
	merged.Symbols = append(merged.Symbols, original.Symbols...)
	for _, s := range updated.Symbols {
		if !containsSymbol(merged.Symbols, s) {
			merged.Symbols = append(merged.Symbols, s)
		}
	}

	// Merge user mentions
	merged.UserMentions = append(merged.UserMentions, original.UserMentions...)
	for _, u := range updated.UserMentions {
		if !containsUserMention(merged.UserMentions, u) {
			merged.UserMentions = append(merged.UserMentions, u)
		}
	}

	// Merge URLs
	merged.Urls = append(merged.Urls, original.Urls...)
	for _, u := range updated.Urls {
		if !containsUrl(merged.Urls, u) {
			merged.Urls = append(merged.Urls, u)
		}
	}

	// Merge media
	var mergedMedia []generated.Media
	if original.Media != nil {
		mergedMedia = append(mergedMedia, *original.Media...)
	}
	if updated.Media != nil {
		for _, m := range *updated.Media {
			if !containsMedia(mergedMedia, m) {
				mergedMedia = append(mergedMedia, m)
			}
		}
	}
	if len(mergedMedia) > 0 {
		merged.Media = &mergedMedia
	}

	// Merge timestamps
	var mergedTimestamps []generated.Timestamp
	if original.Timestamps != nil {
		mergedTimestamps = append(mergedTimestamps, *original.Timestamps...)
	}
	if updated.Timestamps != nil {
		for _, t := range *updated.Timestamps {
			if !containsTimestamp(mergedTimestamps, t) {
				mergedTimestamps = append(mergedTimestamps, t)
			}
		}
	}
	if len(mergedTimestamps) > 0 {
		merged.Timestamps = &mergedTimestamps
	}

	return merged
}

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

// TweetsResult contains tweets and information about whether we've reached the beginning of the thread
type TweetsResult struct {
	Tweets     []*Tweet `json:"tweets"`
	IsComplete bool     `json:"is_complete"` // true if we've reached the beginning of the thread
}

// convertTimelineToTweets converts a generated.Timeline to our TweetsResult struct
func convertTimelineToTweets(timeline *generated.Timeline) (*TweetsResult, error) {
	if timeline == nil {
		return nil, fmt.Errorf("timeline is nil")
	}

	if len(timeline.Instructions) == 0 {
		return nil, fmt.Errorf("no instructions found in timeline")
	}

	var tweets []*Tweet
	isComplete := false

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

	// First, check for TimelineTerminateTimeline instructions
	for _, instruction := range timeline.Instructions {
		terminateTimeline, err := instruction.AsTimelineTerminateTimeline()
		if err == nil && terminateTimeline.Direction == "Top" {
			isComplete = true
			break
		}
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

	return &TweetsResult{
		Tweets:     tweets,
		IsComplete: isComplete,
	}, nil
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
		// Long-form Note Tweet contains additional content and entities
		note := genTweet.NoteTweet.NoteTweetResults.Result

		// Update text with the note tweet content (note tweets usually contain the full text)
		tweet.Text = note.Text

		tweet.Entities = MergeEntities(tweet.Entities, note.EntitySet)

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
