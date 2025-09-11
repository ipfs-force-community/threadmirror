package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

var _ = Describe("MentionService", func() {
	var (
		mentionService *service.MentionService
		db             *sql.DB
		ctx            context.Context
		suite          *testsuit.ContainerTestSuite
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		suite = testsuit.SetupContainerTestSuite(&testing.T{})
		db = suite.DB

		// Create service
		mentionService = service.NewMentionService(
			db,
			&testsuit.MockLLM{},
			&testsuit.MockIPFSStorage{},
		)

		// Reset database for clean test state
		suite.ResetDatabase(&testing.T{})
	})

	AfterEach(func() {
		if suite != nil {
			suite.TearDown(&testing.T{})
		}
	})

	Describe("CreateMention", func() {
		It("should create a new mention successfully", func() {
			userID := "user123"
			threadID := "thread456"
			mentionCreateAt := time.Now()

			result, err := mentionService.CreateMention(ctx, userID, threadID, nil, mentionCreateAt)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.ThreadID).To(Equal(threadID))
		})

		It("should return error for duplicate mention", func() {
			userID := "user123"
			threadID := "thread456"
			mentionCreateAt := time.Now()

			// Create first mention
			_, err := mentionService.CreateMention(ctx, userID, threadID, nil, mentionCreateAt)
			Expect(err).ToNot(HaveOccurred())

			// Try to create duplicate
			_, err = mentionService.CreateMention(ctx, userID, threadID, nil, mentionCreateAt)
			Expect(err).To(Equal(service.ErrMentionAlreadyExists))
		})
	})

	Describe("GetMentions", func() {
		It("should retrieve mentions with pagination", func() {
			userID := "user123"

			// Create multiple mentions
			for i := 0; i < 5; i++ {
				threadID := fmt.Sprintf("thread%d", i)
				_, err := mentionService.CreateMention(ctx, userID, threadID, nil, time.Now())
				Expect(err).ToNot(HaveOccurred())
			}

			// Get mentions with pagination
			mentions, total, err := mentionService.GetMentions(ctx, userID, 3, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(mentions).To(HaveLen(3))
			Expect(total).To(Equal(int64(5)))

			// Test second page
			mentions, total, err = mentionService.GetMentions(ctx, userID, 3, 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(mentions).To(HaveLen(2))
			Expect(total).To(Equal(int64(5)))
		})
	})

	Describe("GetMentionByID", func() {
		It("should retrieve mention by ID", func() {
			userID := "user123"
			threadID := "thread456"
			mentionCreateAt := time.Now()

			// Create mention
			created, err := mentionService.CreateMention(ctx, userID, threadID, nil, mentionCreateAt)
			Expect(err).ToNot(HaveOccurred())

			// Retrieve by ID
			retrieved, err := mentionService.GetMentionByID(ctx, created.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved.ID).To(Equal(created.ID))
			Expect(retrieved.ThreadID).To(Equal(threadID))
		})

		It("should return error for non-existent mention", func() {
			_, err := mentionService.GetMentionByID(ctx, "nonexistent-id")
			Expect(err).To(HaveOccurred())
		})
	})
})
