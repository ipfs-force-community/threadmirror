package testsuit

import (
	"context"
	"io"
	"strings"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs/go-cid"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/datatypes"
)

// MockProcessedMarkRepo is a mock implementation for testing
type MockProcessedMarkRepo struct {
	processedMarks map[string]bool // key:type -> bool
}

func NewMockProcessedMarkRepo() *MockProcessedMarkRepo {
	return &MockProcessedMarkRepo{
		processedMarks: make(map[string]bool),
	}
}

func (m *MockProcessedMarkRepo) makeKey(key string, typ string) string {
	return key + ":" + typ
}

func (m *MockProcessedMarkRepo) IsProcessed(ctx context.Context, key string, typ string) (bool, error) {
	k := m.makeKey(key, typ)
	return m.processedMarks[k], nil
}

func (m *MockProcessedMarkRepo) MarkProcessed(ctx context.Context, key string, typ string) error {
	k := m.makeKey(key, typ)
	m.processedMarks[k] = true
	return nil
}

func (m *MockProcessedMarkRepo) BatchMarkProcessed(ctx context.Context, keys []string, typ string) error {
	for _, key := range keys {
		k := m.makeKey(key, typ)
		m.processedMarks[k] = true
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
		return nil, errutil.ErrNotFound // Simulate no cookies found
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
		return nil, errutil.ErrNotFound
	}
	return mention, nil
}

func (m *MockMentionRepo) GetMentionByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Mention, error) {
	for _, mention := range m.mentions {
		if mention.UserID == userID && mention.ThreadID == threadID {
			return mention, nil
		}
	}
	return nil, errutil.ErrNotFound
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
