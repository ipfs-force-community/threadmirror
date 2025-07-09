package xscraper

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
)

func TestConvertTimelineToTweets(t *testing.T) {
	tests := []struct {
		name     string
		timeline *generated.Timeline
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "nil timeline",
			timeline: nil,
			wantErr:  true,
			errMsg:   "timeline is nil",
		},
		{
			name: "empty instructions",
			timeline: &generated.Timeline{
				Instructions: []generated.InstructionUnion{},
			},
			wantErr: true,
			errMsg:  "no instructions found in timeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tweets, err := convertTimelineToTweets(tt.timeline)
			if tt.wantErr {
				if err == nil {
					t.Errorf("convertTimelineToTweets() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("convertTimelineToTweets() error = %v, want %v", err.Error(), tt.errMsg)
				}
				if tweets != nil {
					t.Errorf("convertTimelineToTweets() expected nil tweets but got %v", tweets)
				}
			} else {
				if err != nil {
					t.Errorf("convertTimelineToTweets() unexpected error = %v", err)
				}
				if tweets == nil {
					t.Errorf("convertTimelineToTweets() expected tweets but got nil")
				}
			}
		})
	}
}

func TestConvertGeneratedTweetToTweet(t *testing.T) {
	tests := []struct {
		name      string
		genTweet  *generated.Tweet
		wantErr   bool
		errMsg    string
		checkFunc func(*testing.T, *Tweet)
	}{
		{
			name:     "nil tweet",
			genTweet: nil,
			wantErr:  true,
			errMsg:   "generated tweet is nil",
		},
		{
			name: "basic tweet conversion",
			genTweet: &generated.Tweet{
				RestId: "1234567890",
				Legacy: &generated.TweetLegacy{
					FullText:          "Hello World! #test",
					ConversationIdStr: "1234567890",
					IsQuoteStatus:     false,
					Lang:              "en",
					CreatedAt:         generated.TwitterTimeFormat("Mon Jan 02 15:04:05 +0000 2006"),
					ReplyCount:        5,
					RetweetCount:      10,
					FavoriteCount:     15,
					QuoteCount:        2,
					BookmarkCount:     1,
					Entities: generated.Entities{
						Hashtags: []generated.Hashtag{
							{"text": "test", "indices": []int{13, 18}},
						},
						Symbols:      []generated.Symbol{},
						Urls:         []generated.Url{},
						UserMentions: []generated.UserMention{},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tweet *Tweet) {
				if tweet.RestID != "1234567890" {
					t.Errorf("Expected RestID '1234567890', got '%s'", tweet.RestID)
				}
				if tweet.Text != "Hello World! #test" {
					t.Errorf("Expected text 'Hello World! #test', got '%s'", tweet.Text)
				}
				if tweet.Lang != "en" {
					t.Errorf("Expected lang 'en', got '%s'", tweet.Lang)
				}
				if tweet.Stats.ReplyCount != 5 {
					t.Errorf("Expected reply count 5, got %d", tweet.Stats.ReplyCount)
				}
				if len(tweet.Entities.Hashtags) != 1 {
					t.Errorf("Expected 1 hashtag, got %d", len(tweet.Entities.Hashtags))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tweet, err := convertGeneratedTweetToTweet(tt.genTweet)
			if tt.wantErr {
				if err == nil {
					t.Errorf("convertGeneratedTweetToTweet() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("convertGeneratedTweetToTweet() error = %v, want %v", err.Error(), tt.errMsg)
				}
				if tweet != nil {
					t.Errorf("convertGeneratedTweetToTweet() expected nil tweet but got %v", tweet)
				}
			} else {
				if err != nil {
					t.Errorf("convertGeneratedTweetToTweet() unexpected error = %v", err)
				}
				if tweet == nil {
					t.Errorf("convertGeneratedTweetToTweet() expected tweet but got nil")
				} else if tt.checkFunc != nil {
					tt.checkFunc(t, tweet)
				}
			}
		})
	}
}

func TestConvertGeneratedUserToUser(t *testing.T) {
	tests := []struct {
		name      string
		genUser   *generated.User
		wantNil   bool
		checkFunc func(*testing.T, *User)
	}{
		{
			name:    "nil user",
			genUser: nil,
			wantNil: true,
		},
		{
			name: "basic user conversion",
			genUser: &generated.User{
				RestId:         "123456789",
				IsBlueVerified: true,
				Legacy: generated.UserLegacy{
					Name:                 "Test User",
					ScreenName:           "testuser",
					Description:          "This is a test user",
					FollowersCount:       1000,
					FriendsCount:         500,
					StatusesCount:        2000,
					Verified:             false,
					CreatedAt:            generated.TwitterTimeFormat("Mon Jan 02 15:04:05 +0000 2006"),
					ProfileImageUrlHttps: "https://example.com/profile.jpg",
				},
			},
			wantNil: false,
			checkFunc: func(t *testing.T, user *User) {
				if user.RestID != "123456789" {
					t.Errorf("Expected RestID '123456789', got '%s'", user.RestID)
				}
				if user.Name != "Test User" {
					t.Errorf("Expected name 'Test User', got '%s'", user.Name)
				}
				if user.ScreenName != "testuser" {
					t.Errorf("Expected screen name 'testuser', got '%s'", user.ScreenName)
				}
				if user.FollowersCount != 1000 {
					t.Errorf("Expected followers count 1000, got %d", user.FollowersCount)
				}
				if !user.IsBlueVerified {
					t.Errorf("Expected IsBlueVerified to be true")
				}
				if user.Verified {
					t.Errorf("Expected Verified to be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := convertGeneratedUserToUser(tt.genUser)
			if tt.wantNil {
				if user != nil {
					t.Errorf("convertGeneratedUserToUser() expected nil but got %v", user)
				}
			} else {
				if user == nil {
					t.Errorf("convertGeneratedUserToUser() expected user but got nil")
				} else if tt.checkFunc != nil {
					tt.checkFunc(t, user)
				}
			}
		})
	}
}

func loadTestTweet(t interface {
	Skipf(format string, args ...any)
	Fatalf(format string, args ...any)
}) *generated.TweetDetailResponse {
	// Read the real tweet data
	data, err := os.ReadFile("testdata/tweet.json")
	if err != nil {
		t.Skipf("Failed to read tweet.json: %v", err)
		return nil
	}

	var response generated.TweetDetailResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal tweet.json: %v", err)
	}
	return &response
}

// TestTweetConversionWithRealData tests conversion using the actual tweet.json file
func TestTweetConversionWithRealData(t *testing.T) {
	response := loadTestTweet(t)
	// Test conversion
	tweetsResult, err := convertTimelineToTweets(response.Data.ThreadedConversationWithInjectionsV2)
	if err != nil {
		t.Fatalf("Failed to convert real tweet data: %v", err)
	}

	// Verify basic fields are populated
	if tweetsResult == nil || len(tweetsResult.Tweets) == 0 {
		t.Fatal("No tweets returned")
	}

	tweet := tweetsResult.Tweets[0] // Get the first tweet

	if tweet.ID == "" {
		t.Error("Tweet ID is empty")
	}

	if tweet.Text == "" {
		t.Error("Tweet text is empty")
	}

	if tweet.CreatedAt.IsZero() {
		t.Error("Tweet creation time is zero")
	}

	// Log some key information for manual verification
	t.Logf("Converted tweet ID: %s", tweet.ID)
	t.Logf("Tweet text length: %d", len(tweet.Text))
	t.Logf("Tweet creation time: %v", tweet.CreatedAt)
	t.Logf("Tweet stats - Replies: %d, Retweets: %d, Likes: %d",
		tweet.Stats.ReplyCount, tweet.Stats.RetweetCount, tweet.Stats.FavoriteCount)

	if tweet.Author != nil {
		t.Logf("Author: %s (@%s)", tweet.Author.Name, tweet.Author.ScreenName)
	}
}

// BenchmarkConvertTimelineToTweets benchmarks the timeline to tweets conversion function
func BenchmarkConvertTimelineToTweets(b *testing.B) {
	// Create a sample response for benchmarking
	sampleTweet := loadTestTweet(b)

	// Create a mock response structure (simplified for benchmarking)
	// Note: This is a simplified version - real benchmarking would need proper mock data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = convertTimelineToTweets(sampleTweet.Data.ThreadedConversationWithInjectionsV2)
	}
}

// TestTimeParsingEdgeCases tests various time format edge cases
func TestTimeParsingEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		timeStr     string
		expectError bool
	}{
		{
			name:        "valid ruby date",
			timeStr:     "Mon Jan 02 15:04:05 +0000 2006",
			expectError: false,
		},
		{
			name:        "invalid format",
			timeStr:     "2006-01-02T15:04:05Z",
			expectError: true,
		},
		{
			name:        "empty string",
			timeStr:     "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genTweet := &generated.Tweet{
				RestId: "123",
				Legacy: &generated.TweetLegacy{
					CreatedAt: generated.TwitterTimeFormat(tc.timeStr),
				},
			}

			tweet, err := convertGeneratedTweetToTweet(genTweet)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tc.expectError {
				if !tweet.CreatedAt.IsZero() {
					t.Errorf("Expected zero time for invalid format, got %v", tweet.CreatedAt)
				}
			} else {
				if tweet.CreatedAt.IsZero() {
					t.Errorf("Expected non-zero time for valid format")
				}
			}
		})
	}
}
