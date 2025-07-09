package xscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestGetTweets(t *testing.T) {
	// Skip this test if Twitter cookies are not available
	cookiesPath := os.Getenv("TWITTER_COOKIES_PATH")
	if cookiesPath == "" {
		cookiesPath = "cookies.json"
	}
	if _, err := os.Stat(cookiesPath); os.IsNotExist(err) {
		t.Skip("Skipping TestGetTweets: Twitter cookies file not found. Set TWITTER_COOKIES_PATH or create cookies.json")
	}

	scraper, err := New(LoginOptions{
		LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
			cookiesBytes, err := os.ReadFile(cookiesPath)
			if err != nil {
				return nil, err
			}
			var cookies []*http.Cookie
			err = json.Unmarshal(cookiesBytes, &cookies)
			if err != nil {
				return nil, err
			}
			return cookies, nil
		},
	}, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	require.NoError(t, err)

	tweetsResult, err := scraper.GetTweets(context.Background(), "1939670365153657119")
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Println("--------------------------------")
	if tweetsBytes, err := json.Marshal(tweetsResult); err == nil {
		fmt.Println(string(tweetsBytes))
	} else {
		t.Fatal(err)
		return
	}

	fmt.Println("--------------------------------")
	for _, tweet := range tweetsResult.Tweets {
		fmt.Println(tweet.Text)
	}
}

func TestCreateTweet(t *testing.T) {
	apiKey := os.Getenv("TWITTER_API_KEY")
	apiKeySecret := os.Getenv("TWITTER_API_KEY_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

	// Skip this test if Twitter API credentials are not available
	if apiKey == "" || apiKeySecret == "" || accessToken == "" || accessTokenSecret == "" {
		t.Skip("Skipping TestCreateTweet: Twitter API credentials not set. Set TWITTER_API_KEY, TWITTER_API_KEY_SECRET, TWITTER_ACCESS_TOKEN, and TWITTER_ACCESS_TOKEN_SECRET")
	}

	// Also check for image path
	imagePath := os.Getenv("TWEET_TEST_IMAGE_PATH")
	if imagePath == "" {
		t.Skip("Skipping TestCreateTweet: TWEET_TEST_IMAGE_PATH not set")
	}

	// Check if image file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Skipping TestCreateTweet: Test image file not found at " + imagePath)
	}

	scraper, err := New(LoginOptions{
		LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
			cookiesPath := os.Getenv("TWITTER_COOKIES_PATH")
			if cookiesPath == "" {
				cookiesPath = "cookies.json"
			}
			cookiesBytes, err := os.ReadFile(cookiesPath)
			if err != nil {
				return nil, err
			}
			var cookies []*http.Cookie
			err = json.Unmarshal(cookiesBytes, &cookies)
			if err != nil {
				return nil, err
			}
			return cookies, nil
		},
		APIKey:            apiKey,
		APIKeySecret:      apiKeySecret,
		AccessToken:       accessToken,
		AccessTokenSecret: accessTokenSecret,
	}, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	require.NoError(t, err)

	ctx := context.Background()
	f, err := os.Open(imagePath)
	require.NoError(t, err)
	r, err := scraper.UploadMedia(ctx, f, int(lo.Must(f.Stat()).Size()))
	require.NoError(t, err)

	tweets, err := scraper.CreateTweet(context.Background(), NewTweet{
		Text:     "test",
		MediaIDs: []string{r.MediaID},
	})
	require.NoError(t, err)

	fmt.Println(tweets)
}
