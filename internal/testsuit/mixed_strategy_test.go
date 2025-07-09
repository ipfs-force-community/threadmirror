package testsuit

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMixedStrategy_MentionService 展示混合测试策略：
// - 外部依赖使用Mock：XScraper（Twitter API）、LLM（AI服务）、IPFS（存储）
// - 数据库使用真实容器：PostgreSQL
func TestMixedStrategy_MentionService(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要testcontainers的测试")
	}

	// 1. 设置真实的数据库环境（testcontainers）
	suite := SetupContainerTestSuite(t)
	defer suite.TearDown(t)

	// 2. 创建Mock外部依赖
	mockXScraper := NewMockXScraper()
	mockLLM := &MockLLM{}
	mockIPFS := &MockIPFSStorage{}

	// 3. 创建真实的Repository（使用testcontainers数据库）
	mentionRepo := sqlrepo.NewMentionRepo(suite.DB)
	threadRepo := sqlrepo.NewThreadRepo(suite.DB)

	// 4. 创建Service（混合使用mock和真实数据库）
	mentionService := service.NewMentionService(
		mentionRepo,
		llm.Model(mockLLM),
		ipfs.Storage(mockIPFS),
		threadRepo,
		suite.DB,
	)

	t.Run("CreateMentionWithMocks", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 配置mock数据 - 创建线程推文和提及推文
		threadTweet := &xscraper.Tweet{
			RestID: "test-thread-tweet-123",
			ID:     "test-thread-tweet-123",
			Text:   "This is the original thread tweet",
			Author: &xscraper.User{
				RestID:          "thread-author-456",
				ScreenName:      "threadauthor",
				Name:            "Thread Author",
				ProfileImageURL: "https://example.com/thread-avatar.jpg",
			},
		}

		mentionTweet := &xscraper.Tweet{
			RestID: "test-mention-tweet-789",
			ID:     "test-mention-tweet-789",
			Text:   "This is a mention reply to @threadauthor",
			Author: &xscraper.User{
				RestID:          "mention-author-123",
				ScreenName:      "mentionuser",
				Name:            "Mention User",
				ProfileImageURL: "https://example.com/mention-avatar.jpg",
			},
		}

		mockXScraper.ClearMockData()
		mockXScraper.AddMockTweet(threadTweet)
		mockXScraper.AddMockTweet(mentionTweet)

		// 创建mention请求（包含线程推文和提及推文）
		req := &service.CreateMentionRequest{
			Tweets: []*xscraper.Tweet{threadTweet, mentionTweet},
		}

		// 测试创建mention（使用真实数据库）
		summary, err := mentionService.CreateMention(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, summary)

		// 验证返回的summary
		assert.Equal(t, mentionTweet.RestID, summary.ID)
		assert.Equal(t, threadTweet.RestID, summary.ThreadID)
		assert.Equal(t, "Thread Author", summary.ThreadAuthor.Name)
		assert.Equal(t, "threadauthor", summary.ThreadAuthor.ScreenName)

		// 验证数据已存储到真实数据库
		retrieved, err := mentionService.GetMentionByID(ctx, summary.ID)
		require.NoError(t, err)
		assert.Equal(t, summary.ID, retrieved.ID)
		assert.Equal(t, summary.ThreadID, retrieved.ThreadID)
	})

	t.Run("TestWithMockErrors", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 配置mock返回错误
		mockXScraper.SetErrorMode(true)

		// 创建基本的推文数据（即使mock会出错，我们仍然可以测试数据库部分）
		threadTweet := &xscraper.Tweet{
			RestID: "error-thread-tweet",
			ID:     "error-thread-tweet",
			Text:   "Error test thread",
			Author: &xscraper.User{
				RestID:     "error-thread-author",
				ScreenName: "errorauthor",
				Name:       "Error Author",
			},
		}

		mentionTweet := &xscraper.Tweet{
			RestID: "error-mention-tweet",
			ID:     "error-mention-tweet",
			Text:   "Error test mention",
			Author: &xscraper.User{
				RestID:     "error-mention-user",
				ScreenName: "erroruser",
				Name:       "Error User",
			},
		}

		req := &service.CreateMentionRequest{
			Tweets: []*xscraper.Tweet{threadTweet, mentionTweet},
		}

		// 数据库操作仍然正常（因为我们的mock错误不影响数据库）
		summary, err := mentionService.CreateMention(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, summary)

		// 验证数据库正常工作
		retrieved, err := mentionService.GetMentionByID(ctx, summary.ID)
		require.NoError(t, err)
		assert.Equal(t, summary.ID, retrieved.ID)

		// 重置mock状态
		mockXScraper.SetErrorMode(false)
	})
}

// TestMixedStrategy_ThreadService 展示ThreadService的混合测试策略
func TestMixedStrategy_ThreadService(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要testcontainers的测试")
	}

	// 设置混合环境
	suite := SetupContainerTestSuite(t)
	defer suite.TearDown(t)

	// Mock外部依赖
	mockIPFS := &MockIPFSStorage{}

	// 真实数据库Repository
	threadRepo := sqlrepo.NewThreadRepo(suite.DB)

	// 创建ThreadService
	threadService := service.NewThreadService(
		threadRepo,
		ipfs.Storage(mockIPFS),
		suite.RedisClient,
		slog.Default(),
	)

	t.Run("GetThreadByID", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 先通过repository创建一个thread（真实数据库）
		thread := &model.Thread{
			ID:        "mixed-test-thread-1",
			Summary:   "Mixed strategy test thread",
			CID:       "QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE7o",
			NumTweets: 3,
		}

		err := threadRepo.CreateThread(ctx, thread)
		require.NoError(t, err)

		// 通过service获取（使用Redis缓存和IPFS mock）
		result, err := threadService.GetThreadByID(ctx, thread.ID)
		require.NoError(t, err)
		assert.Equal(t, thread.ID, result.ID)
		assert.Equal(t, thread.Summary, result.ContentPreview)
	})
}

// TestMockComponents_Isolation 验证各个mock组件的隔离性
func TestMockComponents_Isolation(t *testing.T) {
	t.Run("XScraperMock", func(t *testing.T) {
		mock := NewMockXScraper()

		// 测试正常操作
		ctx := context.Background()
		tweetsResult, err := mock.GetTweets(ctx, "test-id")
		require.NoError(t, err)
		assert.Len(t, tweetsResult.Tweets, 1)

		// 测试错误模式
		mock.SetErrorMode(true)
		_, err = mock.GetTweets(ctx, "test-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Tweet not found")

		// 重置为正常模式
		mock.SetErrorMode(false)
		tweetsResult, err = mock.GetTweets(ctx, "test-id")
		require.NoError(t, err)
		assert.Len(t, tweetsResult.Tweets, 1)
	})

	t.Run("LLMMock", func(t *testing.T) {
		mock := &MockLLM{}
		ctx := context.Background()

		result, err := mock.Call(ctx, "test prompt")
		require.NoError(t, err)
		assert.Equal(t, "Mock AI summary for testing", result)
	})

	t.Run("IPFSMock", func(t *testing.T) {
		mock := &MockIPFSStorage{}
		ctx := context.Background()

		// 测试Get操作
		cid, _ := mock.Add(ctx, nil)
		content, err := mock.Get(ctx, cid)
		require.NoError(t, err)
		defer func() {
			if closeErr := content.Close(); closeErr != nil {
				t.Logf("Warning: Failed to close content: %v", closeErr)
			}
		}()

		// 验证返回的是JSON格式
		buf := make([]byte, 100)
		n, _ := content.Read(buf)
		jsonContent := string(buf[:n])
		assert.Contains(t, jsonContent, "mock-tweet-1")
	})
}
