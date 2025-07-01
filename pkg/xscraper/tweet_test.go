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
	scraper := New(LoginOptions{
		LoadCookies: func(ctx context.Context) ([]*http.Cookie, error) {
			cookiesBytes, err := os.ReadFile("/Users/taoyu/code/threadmirror/cookies.json")
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
