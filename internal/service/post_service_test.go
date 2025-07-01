package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	pkgsql "github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/gorm"
)

type MockPostRepo struct {
	posts map[string]*model.Post
}

func NewMockPostRepo() *MockPostRepo {
	return &MockPostRepo{
		posts: make(map[string]*model.Post),
	}
}

func (m *MockPostRepo) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

func (m *MockPostRepo) CreatePost(ctx context.Context, post *model.Post) error {
	m.posts[post.ID] = post
	return nil
}

func (m *MockPostRepo) GetPosts(ctx context.Context, userID string, limit, offset int) ([]model.Post, int64, error) {
	var result []model.Post
	for _, post := range m.posts {
		if userID == "" || post.UserID == userID {
			result = append(result, *post)
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

func (m *MockPostRepo) GetPostsByUser(ctx context.Context, userID string, limit, offset int) ([]model.Post, int64, error) {
	return m.GetPosts(ctx, userID, limit, offset)
}

type MockLLM struct{}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: "Mock AI summary for testing"}},
	}, nil
}

func (m *MockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "Mock AI summary for testing", nil
}

type MockIPFSStorage struct{}

func (m *MockIPFSStorage) Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error) {
	c, _ := cid.Parse("bafkreiabc123")
	return c, nil
}

func (m *MockIPFSStorage) Get(ctx context.Context, cid cid.Cid) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock content")), nil
}

type MockThreadRepo struct {
	threads      map[string]*model.Thread
	createCalled bool
}

func NewMockThreadRepo() *MockThreadRepo {
	return &MockThreadRepo{
		threads: make(map[string]*model.Thread),
	}
}

func (m *MockThreadRepo) GetThreadByID(id string) (*model.Thread, error) {
	thread, ok := m.threads[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return thread, nil
}

func (m *MockThreadRepo) CreateThread(thread *model.Thread) error {
	m.threads[thread.ID] = thread
	m.createCalled = true
	return nil
}

func (m *MockThreadRepo) GetThreadsByIDs(ids []string) (map[string]*model.Thread, error) {
	result := make(map[string]*model.Thread)
	for _, id := range ids {
		if thread, ok := m.threads[id]; ok {
			result[id] = thread
		}
	}
	return result, nil
}

func TestPostService_CreatePost(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{}
	mockIPFS := &MockIPFSStorage{}

	const testDBFile = "test.db"
	defer os.Remove(testDBFile) //nolint:errcheck
	mockDB, err := pkgsql.New("sqlite", testDBFile, slog.Default())
	assert.NoError(t, err)
	// 自动迁移表结构，先 threads 后 posts，避免外键依赖问题
	err = mockDB.AutoMigrate(&model.Thread{})
	if err != nil {
		t.Fatalf("AutoMigrate Thread failed: %v", err)
	}
	err = mockDB.AutoMigrate(&model.Post{})
	if err != nil {
		t.Fatalf("AutoMigrate Post failed: %v", err)
	}

	postRepo := sqlrepo.NewPostRepo()
	threadRepo := sqlrepo.NewThreadRepo()

	service := NewPostService(
		postRepo,
		mockLLM,
		mockIPFS,
		threadRepo,
	)

	// 构造 tweets
	author := &xscraper.User{
		ID:              "author-id",
		RestID:          "author-restid",
		Name:            "Author Name",
		ScreenName:      "author_screen",
		ProfileImageURL: "http://example.com/profile.jpg",
	}
	tweet1 := &xscraper.Tweet{
		ID:        "tweet-1",
		RestID:    "thread-1",
		Text:      "First tweet",
		CreatedAt: time.Now(),
		Author:    author,
	}
	tweet2 := &xscraper.Tweet{
		ID:        "tweet-2",
		RestID:    "post-1",
		Text:      "Second tweet",
		CreatedAt: time.Now(),
		Author:    author,
	}

	req := &CreatePostRequest{
		Tweets: []*xscraper.Tweet{tweet1, tweet2},
	}

	userID := "user-1"

	// 执行前将 db 注入 ctx
	ctx = pkgsql.WithDBToContext(ctx, mockDB)
	// 执行
	postDetail, err := service.CreateThreadPost(ctx, userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, postDetail)
	assert.Equal(t, tweet2.RestID, postDetail.ID)
	assert.NotNil(t, postDetail.Author)
	assert.Equal(t, author.RestID, postDetail.Author.ID)
	assert.Equal(t, author.Name, postDetail.Author.Name)
	assert.Equal(t, author.ScreenName, postDetail.Author.ScreenName)
	assert.Equal(t, author.ProfileImageURL, postDetail.Author.ProfileImageURL)
}
