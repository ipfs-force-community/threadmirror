package bot

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

type mockJobQueueClient struct{}

func (m *mockJobQueueClient) Enqueue(ctx context.Context, job *jobq.Job) (string, error) {
	return "mock-job-id", nil
}

func createTestBot(t *testing.T) *TwitterBot {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create mock services
	// Mock JobQueueClient
	jobQueueClient := &mockJobQueueClient{}

	return NewTwitterBot(
		[]*xscraper.XScraper{
			{LoginOpts: xscraper.LoginOptions{Username: "testbot"}},
		}, // scrapers
		5*time.Minute, // checkInterval
		jobQueueClient,
		logger,
		"threadmirror", // excludeMentionAuthorPrefix
		"testbot",      // mentionUsername
	)
}

func TestGenerateResponse(t *testing.T) {
	// Skip this test as generateResponse method has been removed
	t.Skip("generateResponse method no longer exists")
}

func TestProcessedMarkService(t *testing.T) {
	// Skip if containers not available
	testsuit.SkipIfContainerUnavailable(t)

	// Setup real database with testcontainers
	suite := testsuit.SetupContainerTestSuite(t)
	defer suite.TearDown(t)

	// Use REAL ProcessedMarkRepo with real database
	processedMarkRepo := sqlrepo.NewProcessedMarkRepo(suite.DB)
	processedMarkService := service.NewProcessedMarkService(processedMarkRepo)
	ctx := context.Background()

	typeVal := "test_type"
	// Test initially not processed
	isProcessed, err := processedMarkService.IsProcessed(ctx, "key1", typeVal)
	assert.NoError(t, err)
	assert.False(t, isProcessed)

	// Mark as processed
	err = processedMarkService.MarkProcessed(ctx, "key1", typeVal)
	assert.NoError(t, err)

	isProcessed, err = processedMarkService.IsProcessed(ctx, "key1", typeVal)
	assert.NoError(t, err)
	assert.True(t, isProcessed)

	// Test different key
	isProcessed, err = processedMarkService.IsProcessed(ctx, "key2", typeVal)
	assert.NoError(t, err)
	assert.False(t, isProcessed)

	// Test different type, same key
	isProcessed, err = processedMarkService.IsProcessed(ctx, "key1", "other_type")
	assert.NoError(t, err)
	assert.False(t, isProcessed)
}

func TestBotCookieService(t *testing.T) {
	// Skip if containers not available
	testsuit.SkipIfContainerUnavailable(t)

	// Setup real database with testcontainers
	suite := testsuit.SetupContainerTestSuite(t)
	defer suite.TearDown(t)

	// Use REAL BotCookieRepo with real database
	botCookieRepo := sqlrepo.NewBotCookieRepo(suite.DB)
	botCookieService := service.NewBotCookieService(botCookieRepo)
	ctx := context.Background()
	testEmail := "test@example.com"
	testUsername := "testbot"

	// Test initially no cookies - should return errutil.ErrNotFound
	cookies, err := botCookieService.LoadCookies(ctx, testEmail, testUsername)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errutil.ErrNotFound)
	assert.Nil(t, cookies)

	// Save some test cookies
	testCookies := []*http.Cookie{
		{
			Name:  "test_cookie",
			Value: "test_value",
		},
	}
	err = botCookieService.SaveCookies(ctx, testEmail, testUsername, testCookies)
	assert.NoError(t, err)

	// Load cookies - should now return the saved cookies
	cookies, err = botCookieService.LoadCookies(ctx, testEmail, testUsername)
	assert.NoError(t, err)
	assert.NotNil(t, cookies)
	assert.Len(t, cookies, 1)
	assert.Equal(t, "test_cookie", cookies[0].Name)
	assert.Equal(t, "test_value", cookies[0].Value)
}

func TestGetStats(t *testing.T) {
	bot := createTestBot(t)

	stats := bot.GetStats()

	assert.Equal(t, true, stats["enabled"])
	assert.Equal(t, "testbot", stats["mention_username"])
	assert.Equal(t, "5m0s", stats["check_interval"])
	assert.Equal(t, true, stats["randomized"])
	assert.Equal(t, "database", stats["storage_type"])
}

func TestRandomizedInterval(t *testing.T) {
	bot := createTestBot(t)
	baseInterval := bot.checkInterval // 5 minutes

	// Test multiple calculations to ensure randomization
	intervals := make([]time.Duration, 20)
	for i := 0; i < 20; i++ {
		intervals[i] = bot.randomizedInterval()
	}

	// Calculate expected range (±30% of base interval)
	jitterRange := time.Duration(float64(baseInterval) * 0.3)
	minExpected := baseInterval - jitterRange
	maxExpected := baseInterval + jitterRange

	// Ensure minimum of 30 seconds
	if minExpected < 30*time.Second {
		minExpected = 30 * time.Second
	}

	for _, interval := range intervals {
		// Should be within expected bounds
		assert.GreaterOrEqual(t, interval, minExpected,
			"Interval should be >= min expected (%v)", minExpected)
		assert.LessOrEqual(t, interval, maxExpected,
			"Interval should be <= max expected (%v)", maxExpected)
	}

	// Check that we get variation (not all the same)
	allSame := true
	for i := 1; i < len(intervals); i++ {
		if intervals[i] != intervals[0] {
			allSame = false
			break
		}
	}
	assert.False(t, allSame, "Should generate different intervals, got all same: %v", intervals[0])
}
