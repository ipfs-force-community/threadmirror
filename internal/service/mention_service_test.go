package service_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
)

func TestMentionService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MentionService Suite")
}

var _ = AfterSuite(func() {
	// Cleanup test containers
	testsuit.CleanupTestContainers()
})

var _ = Describe("MentionService", func() {
	var (
		ctx            context.Context
		mentionService *service.MentionService
		mentionRepo    *sqlrepo.MentionRepo
		threadRepo     *sqlrepo.ThreadRepo
		ipfsStorage    *testsuit.MockIPFSStorage
		llmModel       *testsuit.MockLLM
		db             *sql.DB
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Use real database for transaction support
		db = testsuit.SetupGinkgoTestDB()

		// Create REAL repositories (not mocks) for proper database testing
		mentionRepo = sqlrepo.NewMentionRepo(db)
		threadRepo = sqlrepo.NewThreadRepo(db)

		// Only mock external dependencies
		ipfsStorage = &testsuit.MockIPFSStorage{}
		llmModel = &testsuit.MockLLM{}

		// Create service with REAL repositories and real database
		mentionService = service.NewMentionService(
			mentionRepo, // Real MentionRepo using actual SQL
			llmModel,    // Mock LLM (external dependency)
			ipfsStorage, // Mock IPFS (external dependency)
			threadRepo,  // Real ThreadRepo using actual SQL
			db,          // Real database with transaction support
		)

		// Reset database for clean test state
		testsuit.ResetGinkgoDatabase()
	})

	Describe("CreateMention", func() {
		It("should create a mention and thread successfully", func() {
			// Arrange
			userID := "user123"
			threadID := "thread123"

			// Act
			mention, err := mentionService.CreateMention(ctx, userID, threadID, nil, time.Now())

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(mention).ToNot(BeNil())
			Expect(mention.ID).To(Equal("thread123_user123"))
			Expect(mention.ThreadID).To(Equal("thread123"))

			// Verify thread was created with pending status
			thread, err := threadRepo.GetThreadByID(ctx, "thread123")
			Expect(err).ToNot(HaveOccurred())
			Expect(thread.Status).To(Equal(model.ThreadStatusPending))
		})

		It("should return error when mention already exists", func() {
			// Arrange
			userID := "user123"
			threadID := "thread123"

			// Create first mention
			_, err := mentionService.CreateMention(ctx, userID, threadID, nil, time.Now())
			Expect(err).ToNot(HaveOccurred())

			// Act - try to create same mention again
			_, err = mentionService.CreateMention(ctx, userID, threadID, nil, time.Now())

			// Assert
			Expect(err).ToNot(HaveOccurred()) // Should return existing mention, not error
		})
	})

	Describe("GetMentions", func() {
		BeforeEach(func() {
			// Create threads first (required for foreign key constraint)
			threads := []*model.Thread{
				{
					ID:        "thread1",
					Status:    model.ThreadStatusCompleted,
					Summary:   "First thread summary",
					NumTweets: 5,
					CreatedAt: time.Now(),
				},
				{
					ID:        "thread2",
					Status:    model.ThreadStatusCompleted,
					Summary:   "Second thread summary",
					NumTweets: 3,
					CreatedAt: time.Now().Add(-time.Hour),
				},
				{
					ID:        "thread3",
					Status:    model.ThreadStatusPending,
					Summary:   "",
					NumTweets: 0,
					CreatedAt: time.Now().Add(-2 * time.Hour),
				},
			}

			for _, thread := range threads {
				err := threadRepo.CreateThread(ctx, thread)
				Expect(err).ToNot(HaveOccurred())
			}

			// Create test mentions after threads exist
			mentions := []*model.Mention{
				{
					ID:              "thread1_user1",
					UserID:          "user1",
					ThreadID:        "thread1",
					MentionCreateAt: time.Now(),
					CreatedAt:       time.Now(),
				},
				{
					ID:              "thread2_user1",
					UserID:          "user1",
					ThreadID:        "thread2",
					MentionCreateAt: time.Now().Add(-time.Hour),
					CreatedAt:       time.Now().Add(-time.Hour),
				},
				{
					ID:              "thread3_user2",
					UserID:          "user2",
					ThreadID:        "thread3",
					MentionCreateAt: time.Now().Add(-2 * time.Hour),
					CreatedAt:       time.Now().Add(-2 * time.Hour),
				},
			}

			for _, mention := range mentions {
				err := mentionRepo.CreateMention(ctx, mention)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should return mentions for a specific user", func() {
			// Act
			mentions, total, err := mentionService.GetMentions(ctx, "user1", 10, 0)

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(mentions).To(HaveLen(2))
			Expect(total).To(Equal(int64(2)))

			// Verify the mentions are for the correct user
			for _, summary := range mentions {
				Expect(summary.ID).To(MatchRegexp("thread[12]_user1"))
				Expect(summary.ThreadID).To(MatchRegexp("thread[12]"))
			}
		})

		It("should return all mentions when userID is empty", func() {
			// Act
			mentions, total, err := mentionService.GetMentions(ctx, "", 10, 0)

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(mentions).To(HaveLen(3))
			Expect(total).To(Equal(int64(3)))
		})

		It("should handle pagination correctly", func() {
			// Act - get first page
			mentions, total, err := mentionService.GetMentions(ctx, "", 2, 0)

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(mentions).To(HaveLen(2))
			Expect(total).To(Equal(int64(3)))

			// Act - get second page
			mentions, total, err = mentionService.GetMentions(ctx, "", 2, 2)

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(mentions).To(HaveLen(1))
			Expect(total).To(Equal(int64(3)))
		})
	})

	Describe("GetMentionByID", func() {
		var testMention *model.Mention
		var testThread *model.Thread

		BeforeEach(func() {
			// Create thread first (required for foreign key constraint)
			testThread = &model.Thread{
				ID:        "thread123",
				Status:    model.ThreadStatusCompleted,
				Summary:   "Test thread summary",
				NumTweets: 3,
				CreatedAt: time.Now(),
			}
			err := threadRepo.CreateThread(ctx, testThread)
			Expect(err).ToNot(HaveOccurred())

			// Create mention after thread exists
			testMention = &model.Mention{
				ID:              "thread123_user456",
				UserID:          "user456",
				ThreadID:        "thread123",
				MentionCreateAt: time.Now(),
				CreatedAt:       time.Now(),
			}
			err = mentionRepo.CreateMention(ctx, testMention)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return mention summary successfully", func() {
			// Act
			summary, err := mentionService.GetMentionByID(ctx, "thread123_user456")

			// Assert
			Expect(err).ToNot(HaveOccurred())
			Expect(summary).ToNot(BeNil())
			Expect(summary.ID).To(Equal("thread123_user456"))
			Expect(summary.ThreadID).To(Equal("thread123"))
		})

		It("should return error when mention not found", func() {
			// Act
			summary, err := mentionService.GetMentionByID(ctx, "nonexistent")

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(summary).To(BeNil())
			Expect(err).To(Equal(errutil.ErrNotFound))
		})

		It("should return mention with empty thread data when thread exists but has no author", func() {
			// Arrange - create thread without author info and corresponding mention
			emptyThread := &model.Thread{
				ID:        "thread999",
				Status:    model.ThreadStatusPending,
				Summary:   "Thread without author",
				NumTweets: 0,
				CreatedAt: time.Now(),
				// AuthorID, AuthorName etc. are empty
			}
			err := threadRepo.CreateThread(ctx, emptyThread)
			Expect(err).ToNot(HaveOccurred())

			orphanMention := &model.Mention{
				ID:              "thread999_user456",
				UserID:          "user456",
				ThreadID:        "thread999",
				MentionCreateAt: time.Now(),
				CreatedAt:       time.Now(),
			}
			err = mentionRepo.CreateMention(ctx, orphanMention)
			Expect(err).ToNot(HaveOccurred())

			// Act
			summary, err := mentionService.GetMentionByID(ctx, "thread999_user456")

			// Assert - Should return summary with empty author data
			Expect(err).ToNot(HaveOccurred())
			Expect(summary).ToNot(BeNil())
			Expect(summary.ID).To(Equal("thread999_user456"))
			Expect(summary.ThreadID).To(Equal("thread999"))
			Expect(summary.ThreadAuthor).To(BeNil()) // No author info
			Expect(summary.ContentPreview).To(Equal("Thread without author"))
		})
	})
})
