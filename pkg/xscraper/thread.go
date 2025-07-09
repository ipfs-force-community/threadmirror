package xscraper

import (
	"context"
	"fmt"
	"sort"
)

// GetCompleteThread 获取完整的推文串
// 这是一个纯函数，会继续调用GetTweets直到IsComplete为true或达到最大尝试次数
func GetCompleteThread(ctx context.Context, scraper XScraperInterface, tweetID string, maxAttempts int) ([]*Tweet, error) {
	if maxAttempts <= 0 {
		maxAttempts = 10 // 默认最大尝试次数
	}

	var allTweets []*Tweet
	seenTweetIDs := make(map[string]bool) // 去重
	attempts := 0

	for attempts < maxAttempts {
		attempts++

		tweetsResult, err := scraper.GetTweets(ctx, tweetID)
		if err != nil {
			return nil, fmt.Errorf("attempt %d failed: %w", attempts, err)
		}

		// 添加新推文（去重）
		for _, tweet := range tweetsResult.Tweets {
			if tweet.RestID != "" && !seenTweetIDs[tweet.RestID] {
				allTweets = append(allTweets, tweet)
				seenTweetIDs[tweet.RestID] = true
			}
		}

		// 如果已完整，返回结果
		if tweetsResult.IsComplete || len(tweetsResult.Tweets) == 0 {
			break
		}

		// 若未完整，则尝试获取最早一条推文的父级推文 ID，用于下一次查询
		oldestTweet := tweetsResult.Tweets[0]

		// 无法找到更早的推文，终止循环
		if oldestTweet == nil || !oldestTweet.IsReply {
			break
		}

		// 如果父级推文已获取过，则终止，避免死循环
		if seenTweetIDs[oldestTweet.InReplyToStatusID] {
			break
		}

		// 更新 tweetID，用于下一轮查询
		tweetID = oldestTweet.InReplyToStatusID
	}

	// 使用 RestID 升序排序后返回（RestID 越小代表越早的推文）
	sort.Slice(allTweets, func(i, j int) bool {
		return allTweets[i].RestID < allTweets[j].RestID
	})

	return allTweets, nil
}
