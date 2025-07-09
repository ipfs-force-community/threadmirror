package testsuit

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMixedStrategy 展示了如何在集成测试中混合使用真实和Mock组件
// 这是一个最佳实践示例，展示了：
// 1. 使用真实数据库进行Repository和Service测试
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

	// 真实数据库Repository
	threadRepo := sqlrepo.NewThreadRepo(suite.DB)

	// 创建ThreadService
	threadService := service.NewThreadService(
		threadRepo,
		ipfs.Storage(mockIPFS),
		mockLLM,
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
			Status:    model.ThreadStatusCompleted,
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
