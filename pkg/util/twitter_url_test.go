package util

import (
	"testing"
)

func TestParseTwitterURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectValid bool
		expectID    string
		expectUser  string
		expectError bool
	}{
		// 有效的推特链接
		{
			name:        "Standard Twitter URL",
			url:         "https://twitter.com/elonmusk/status/1234567890123456789",
			expectValid: true,
			expectID:    "1234567890123456789",
			expectUser:  "elonmusk",
			expectError: false,
		},
		{
			name:        "X.com URL",
			url:         "https://x.com/elonmusk/status/1234567890123456789",
			expectValid: true,
			expectID:    "1234567890123456789",
			expectUser:  "elonmusk",
			expectError: false,
		},
		{
			name:        "Mobile Twitter URL",
			url:         "https://mobile.twitter.com/username/status/9876543210987654321",
			expectValid: true,
			expectID:    "9876543210987654321",
			expectUser:  "username",
			expectError: false,
		},
		{
			name:        "URL with query parameters",
			url:         "https://twitter.com/test_user/status/1111111111111111111?s=20&t=abc123",
			expectValid: true,
			expectID:    "1111111111111111111",
			expectUser:  "test_user",
			expectError: false,
		},
		{
			name:        "URL without protocol",
			url:         "twitter.com/user123/status/2222222222222222222",
			expectValid: true,
			expectID:    "2222222222222222222",
			expectUser:  "user123",
			expectError: false,
		},
		{
			name:        "Statuses instead of status",
			url:         "https://twitter.com/username/statuses/3333333333333333333",
			expectValid: true,
			expectID:    "3333333333333333333",
			expectUser:  "username",
			expectError: false,
		},
		// 无效的链接
		{
			name:        "Empty URL",
			url:         "",
			expectValid: false,
			expectError: true,
		},
		{
			name:        "Invalid domain",
			url:         "https://facebook.com/user/status/1234567890",
			expectValid: false,
			expectError: true,
		},
		{
			name:        "Missing status path",
			url:         "https://twitter.com/username",
			expectValid: false,
			expectError: true,
		},
		{
			name:        "Invalid tweet ID (too short)",
			url:         "https://twitter.com/username/status/123",
			expectValid: false,
			expectError: true,
		},
		{
			name:        "Invalid tweet ID (too long)",
			url:         "https://twitter.com/username/status/12345678901234567890123456789",
			expectValid: false,
			expectError: true,
		},
		{
			name:        "Non-numeric tweet ID",
			url:         "https://twitter.com/username/status/abcdefghij",
			expectValid: false,
			expectError: true,
		},
		{
			name:        "Malformed URL",
			url:         "not-a-url",
			expectValid: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTwitterURL(tt.url)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expectError {
				return // 如果期望错误，不需要继续检查结果
			}

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}

			if tt.expectValid {
				if result.TweetID != tt.expectID {
					t.Errorf("Expected TweetID=%s, got %s", tt.expectID, result.TweetID)
				}

				if result.Username != tt.expectUser {
					t.Errorf("Expected Username=%s, got %s", tt.expectUser, result.Username)
				}
			}
		})
	}
}

func TestExtractTweetID(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectID    string
		expectError bool
	}{
		{
			name:        "Valid URL",
			url:         "https://twitter.com/test/status/1234567890123456789",
			expectID:    "1234567890123456789",
			expectError: false,
		},
		{
			name:        "Invalid URL",
			url:         "https://invalid.com/test",
			expectID:    "",
			expectError: true,
		},
		{
			name:        "Empty URL",
			url:         "",
			expectID:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractTweetID(tt.url)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expectID {
				t.Errorf("Expected TweetID=%s, got %s", tt.expectID, result)
			}
		})
	}
}
