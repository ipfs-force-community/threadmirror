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
	}, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	require.NoError(t, err)

	tweets, err := scraper.GetTweets(context.Background(), "1939670365153657119")
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Println("--------------------------------")
	if tweetsBytes, err := json.Marshal(tweets); err == nil {
		fmt.Println(string(tweetsBytes))
	} else {
		t.Fatal(err)
		return
	}

	fmt.Println("--------------------------------")
	for _, tweet := range tweets {
		fmt.Println(tweet.Text)
	}
}

func TestCreateTweet(t *testing.T) {
	apiKey := os.Getenv("TWITTER_API_KEY")
	apiKeySecret := os.Getenv("TWITTER_API_KEY_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

	require.NotEmpty(t, apiKey, "TWITTER_API_KEY must be set")
	require.NotEmpty(t, apiKeySecret, "TWITTER_API_KEY_SECRET must be set")
	require.NotEmpty(t, accessToken, "TWITTER_ACCESS_TOKEN must be set")
	require.NotEmpty(t, accessTokenSecret, "TWITTER_ACCESS_TOKEN_SECRET must be set")

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
	imagePath := os.Getenv("TWEET_TEST_IMAGE_PATH")
	require.NotEmpty(t, imagePath, "TWEET_TEST_IMAGE_PATH must be set")
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
