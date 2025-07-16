package util

import (
	"strings"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

// ExtractTweetTexts converts scraped tweets to simple TweetText format
// Uses Tweet.GetDisplayableText() to get clean text without entities
func ExtractTweetTexts(tweets []*xscraper.Tweet) []model.TweetText {
	if len(tweets) == 0 {
		return []model.TweetText{}
	}

	tweetTexts := make([]model.TweetText, 0, len(tweets))

	for _, tweet := range tweets {
		displayableText := tweet.GetDisplayableText()
		if strings.TrimSpace(displayableText) == "" {
			continue // Skip empty tweets
		}

		tweetText := model.TweetText{
			TweetID:         tweet.RestID,
			DisplayableText: displayableText,
			TranslatedText:  "", // Will be filled during translation
		}

		tweetTexts = append(tweetTexts, tweetText)
	}

	return tweetTexts
}
