package testsuit

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMixedStrategy 展示了如何在集成测试中混合使用真实和Mock组件
// 这是一个最佳实践示例，展示了：
// 1. 使用真实数据库进行Service测试（SQL First架构）
// 2. Mock外部依赖（LLM, IPFS, XScraper）
// 3. 验证组件之间的集成

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
	mockLLM := &MockLLM{}

	// 创建ThreadService（SQL First架构，直接使用数据库）
	threadService := service.NewThreadService(
		suite.DB,
		ipfs.Storage(mockIPFS),
		mockLLM,
		suite.RedisClient,
		slog.Default(),
	)

	t.Run("GetThreadByID_NotFound", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 测试获取不存在的thread
		_, err := threadService.GetThreadByID(ctx, "nonexistent-thread-id")
		require.Error(t, err)
		assert.Equal(t, service.ErrThreadNotFound, err)
	})

	t.Run("UpdateThreadStatus", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 测试更新不存在thread的状态
		err := threadService.UpdateThreadStatus(ctx, "nonexistent-thread-id", "completed", 1)
		require.Error(t, err)
	})

	t.Run("GetRetryThreads", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 测试获取重试线程（应该返回空列表）
		threads, err := threadService.GetStuckScrapingThreadsForRetry(ctx, 3600, 3)
		require.NoError(t, err)
		assert.Empty(t, threads)

		threads, err = threadService.GetOldPendingThreadsForRetry(ctx, 3600, 3)
		require.NoError(t, err)
		assert.Empty(t, threads)

		threads, err = threadService.GetFailedThreadsForRetry(ctx, 3600, 3)
		require.NoError(t, err)
		assert.Empty(t, threads)
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

// TestServiceIntegration 测试服务层集成
func TestServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要testcontainers的测试")
	}

	suite := SetupContainerTestSuite(t)
	defer suite.TearDown(t)

	t.Run("MentionService_Integration", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 测试MentionService的基本功能
		mentionService := suite.MentionService
		require.NotNil(t, mentionService)

		// 测试创建mention（会自动创建thread）
		userID := "test-user-123"
		threadID := "test-thread-456"

		mention, err := mentionService.CreateMention(ctx, userID, threadID, nil, time.Now())
		require.NoError(t, err)
		assert.NotNil(t, mention)
		assert.Equal(t, threadID, mention.ThreadID)

		// 测试获取mentions
		mentions, total, err := mentionService.GetMentions(ctx, userID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, mentions, 1)
	})

	t.Run("ThreadService_Integration", func(t *testing.T) {
		ctx := context.Background()
		suite.ResetDatabase(t)

		// 测试ThreadService的基本功能
		threadService := suite.ThreadService
		require.NotNil(t, threadService)

		// 测试获取不存在的thread
		_, err := threadService.GetThreadByID(ctx, "nonexistent")
		assert.Equal(t, service.ErrThreadNotFound, err)
	})
}
