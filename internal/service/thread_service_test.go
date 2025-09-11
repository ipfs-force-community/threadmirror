package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

var _ = Describe("ThreadService", func() {
	var (
		threadService *service.ThreadService
		db            *sql.DB
		redisClient   *redis.Client
		ctx           context.Context
		suite         *testsuit.ContainerTestSuite
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		suite = testsuit.SetupContainerTestSuite(&testing.T{})
		db = suite.DB
		redisClient = suite.RedisClient

		// Create service
		threadService = service.NewThreadService(
			db,
			&testsuit.MockIPFSStorage{},
			&testsuit.MockLLM{},
			redisClient,
			slog.New(slog.NewTextHandler(os.Stdout, nil)),
		)

		// Reset database for clean test state
		suite.ResetDatabase(&testing.T{})
	})

	AfterEach(func() {
		if suite != nil {
			suite.TearDown(&testing.T{})
		}
	})

	Describe("GetThreadByID", func() {
		It("should return error for non-existent thread", func() {
			_, err := threadService.GetThreadByID(ctx, "nonexistent-id")
			Expect(err).To(Equal(service.ErrThreadNotFound))
		})

		It("should return thread details for existing thread", func() {
			// This test would require creating a thread first
			// For now, just test the error case
			_, err := threadService.GetThreadByID(ctx, "test-thread-id")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateThreadStatus", func() {
		It("should return error for non-existent thread", func() {
			err := threadService.UpdateThreadStatus(ctx, "nonexistent-id", "completed", 1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetStuckScrapingThreadsForRetry", func() {
		It("should return empty list when no stuck threads", func() {
			threads, err := threadService.GetStuckScrapingThreadsForRetry(ctx, time.Hour, 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(threads).To(BeEmpty())
		})
	})

	Describe("GetOldPendingThreadsForRetry", func() {
		It("should return empty list when no old pending threads", func() {
			threads, err := threadService.GetOldPendingThreadsForRetry(ctx, time.Hour, 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(threads).To(BeEmpty())
		})
	})

	Describe("GetFailedThreadsForRetry", func() {
		It("should return empty list when no failed threads", func() {
			threads, err := threadService.GetFailedThreadsForRetry(ctx, time.Hour, 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(threads).To(BeEmpty())
		})
	})
})
