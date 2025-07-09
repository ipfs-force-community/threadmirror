package util

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

// TwitterURL 表示解析后的推特链接信息
type TwitterURL struct {
	TweetID  string // 推文ID
	Username string // 用户名
	IsValid  bool   // 是否是有效的推特链接
}

// ParseTwitterURL 解析推特链接并提取相关信息
// 支持的URL格式:
// - https://twitter.com/username/status/1234567890
// - https://x.com/username/status/1234567890
// - https://mobile.twitter.com/username/status/1234567890
// - https://www.twitter.com/username/status/1234567890
// - 带查询参数的链接: https://twitter.com/username/status/1234567890?s=20&t=abc123
func ParseTwitterURL(twitterURL string) (*TwitterURL, error) {
	if twitterURL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	// 清理URL，移除多余的空格
	twitterURL = strings.TrimSpace(twitterURL)

	// 如果没有协议，添加https://
	if !strings.HasPrefix(twitterURL, "http://") && !strings.HasPrefix(twitterURL, "https://") {
		twitterURL = "https://" + twitterURL
	}

	// 解析URL
	parsedURL, err := url.Parse(twitterURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	// 检查是否是推特域名
	hostname := strings.ToLower(parsedURL.Hostname())
	validDomains := []string{
		"twitter.com",
		"x.com",
		"mobile.twitter.com",
		"www.twitter.com",
		"mobile.x.com",
		"www.x.com",
	}

	isValidDomain := slices.Contains(validDomains, hostname)

	if !isValidDomain {
		return &TwitterURL{IsValid: false}, fmt.Errorf("not a valid Twitter/X domain: %s", hostname)
	}

	// 使用正则表达式匹配路径
	// 支持格式: /username/status/tweetid 或 /username/statuses/tweetid
	pathRegex := regexp.MustCompile(`^/([^/]+)/status(?:es)?/(\d+)`)
	matches := pathRegex.FindStringSubmatch(parsedURL.Path)

	if len(matches) != 3 {
		return &TwitterURL{IsValid: false}, fmt.Errorf("invalid Twitter URL path format")
	}

	username := matches[1]
	tweetID := matches[2]

	// 验证tweet ID（应该是数字且合理长度）
	if len(tweetID) < 10 || len(tweetID) > 25 {
		return &TwitterURL{IsValid: false}, fmt.Errorf("invalid tweet ID format")
	}

	return &TwitterURL{
		TweetID:  tweetID,
		Username: username,
		IsValid:  true,
	}, nil
}

// ExtractTweetID 简化函数，只提取tweet ID
func ExtractTweetID(twitterURL string) (string, error) {
	parsed, err := ParseTwitterURL(twitterURL)
	if err != nil {
		return "", err
	}

	if !parsed.IsValid {
		return "", fmt.Errorf("invalid Twitter URL")
	}

	return parsed.TweetID, nil
}
