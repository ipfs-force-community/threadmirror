package testsuit

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs/go-cid"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/datatypes"
)

// MockProcessedMentionRepo is a mock implementation for testing
type MockProcessedMentionRepo struct {
	processedMentions map[string]bool // userID:tweetID -> bool
}

func NewMockProcessedMentionRepo() *MockProcessedMentionRepo {
	return &MockProcessedMentionRepo{
		processedMentions: make(map[string]bool),
	}
}

func (m *MockProcessedMentionRepo) makeKey(userID string, tweetID string) string {
	return userID + ":" + tweetID
}

func (m *MockProcessedMentionRepo) IsProcessed(ctx context.Context, userID string, tweetID string) (bool, error) {
	key := m.makeKey(userID, tweetID)
	return m.processedMentions[key], nil
}

func (m *MockProcessedMentionRepo) MarkProcessed(ctx context.Context, userID string, tweetID string) error {
	key := m.makeKey(userID, tweetID)
	m.processedMentions[key] = true
	return nil
}

func (m *MockProcessedMentionRepo) BatchMarkProcessed(ctx context.Context, userID string, tweetIDs []string) error {
	for _, tweetID := range tweetIDs {
		key := m.makeKey(userID, tweetID)
		m.processedMentions[key] = true
	}
	return nil
}

// MockBotCookieRepo is a mock implementation for testing
type MockBotCookieRepo struct {
	cookies map[string][]byte // email:username -> JSON data
}

func NewMockBotCookieRepo() *MockBotCookieRepo {
	return &MockBotCookieRepo{
		cookies: make(map[string][]byte),
	}
}

func (m *MockBotCookieRepo) makeKey(email, username string) string {
	return email + ":" + username
}

func (m *MockBotCookieRepo) GetCookies(ctx context.Context, email, username string) (datatypes.JSON, error) {
	key := m.makeKey(email, username)
	cookies, exists := m.cookies[key]
	if !exists {
		return nil, nil // Simulate no cookies found
	}
	return datatypes.JSON(cookies), nil
}

func (m *MockBotCookieRepo) SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error {
	// This would normally marshal the data in the real repo
	key := m.makeKey(email, username)
	m.cookies[key] = []byte(`[]`) // Store empty JSON for testing
	return nil
}

// MockLLM is a mock implementation for testing
type MockLLM struct{}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Return a simple mock response
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "Mock AI summary for testing",
			},
		},
	}, nil
}

func (m *MockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "Mock AI summary for testing", nil
}

// MockIPFSStorage is a mock implementation for testing
type MockIPFSStorage struct{}

func (m *MockIPFSStorage) Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error) {
	// Return a fixed CID for testing
	c, _ := cid.Parse("bafkreiabc123")
	return c, nil
}

func (m *MockIPFSStorage) Get(ctx context.Context, cid cid.Cid) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock content")), nil
}

// MockMentionRepo is a mock implementation for MentionRepoInterface
// Stores mentions in memory for testing
type MockMentionRepo struct {
	mentions map[string]*model.Mention
}

func NewMockMentionRepo() *MockMentionRepo {
	return &MockMentionRepo{
		mentions: make(map[string]*model.Mention),
	}
}

func (m *MockMentionRepo) GetMentionByID(ctx context.Context, id string) (*model.Mention, error) {
	mention, ok := m.mentions[id]
	if !ok {
		return nil, fmt.Errorf("mention not found")
	}
	return mention, nil
}

func (m *MockMentionRepo) GetMentionByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Mention, error) {
	for _, mention := range m.mentions {
		if mention.UserID == userID && mention.ThreadID == threadID {
			return mention, nil
		}
	}
	return nil, nil
}

func (m *MockMentionRepo) CreateMention(ctx context.Context, mention *model.Mention) error {
	m.mentions[mention.ID] = mention
	return nil
}

func (m *MockMentionRepo) GetMentions(ctx context.Context, userID string, limit, offset int) ([]model.Mention, int64, error) {
	var result []model.Mention
	for _, mention := range m.mentions {
		if userID == "" || mention.UserID == userID {
			result = append(result, *mention)
		}
	}
	total := int64(len(result))
	if offset > len(result) {
		offset = len(result)
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *MockMentionRepo) GetMentionsByUser(ctx context.Context, userID string, limit, offset int) ([]model.Mention, int64, error) {
	return m.GetMentions(ctx, userID, limit, offset)
}
