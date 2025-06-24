package xscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
)

func TestGetTweets(t *testing.T) {
	cookiesFile := os.Getenv("COOKIES_FILE")

	scraper := New(LoginOptions{
		LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
			cookiesBytes, err := os.ReadFile(cookiesFile)
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
		SaveCookies: func(ctx context.Context, cookies []*http.Cookie) error {
			cookiesBytes, err := json.Marshal(cookies)
			if err != nil {
				return err
			}
			return os.WriteFile(cookiesFile, cookiesBytes, 0644)
		},
		Username: os.Getenv("X_USERNAME"),
		Password: os.Getenv("X_PASSWORD"),
		Email:    os.Getenv("X_EMAIL"),
	}, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	tweets, err := scraper.GetTweets(context.Background(), "1665301166576799744")
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
