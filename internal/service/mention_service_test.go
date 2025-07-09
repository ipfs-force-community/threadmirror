package service_test

import (
	"context"
	"errors"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = AfterSuite(func() {
	// Clean up all test containers after all tests complete
	testsuit.CleanupTestContainers()
})

var _ = Describe("MentionService", func() {
	var (
		mentionService *service.MentionService
		mentionRepo    service.MentionRepoInterface
		threadRepo     service.ThreadRepoInterface
		mockLLM        *testsuit.MockLLM
		mockIPFS       *testsuit.MockIPFSStorage
		db             *sql.DB
		ctx            context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		db = testsuit.SetupGinkgoTestDB()

		// Clean database before each test
		testsuit.ResetGinkgoDatabase()

		// Initialize repos with database
		mentionRepo = sqlrepo.NewMentionRepo(db)
		threadRepo = sqlrepo.NewThreadRepo(db)

		// Initialize mocks
		mockLLM = &testsuit.MockLLM{}
		mockIPFS = &testsuit.MockIPFSStorage{}

		// Create service with mocks
		mentionService = service.NewMentionService(
			mentionRepo,
			mockLLM,
			mockIPFS,
			threadRepo,
			db,
		)
	})

	Describe("CreateMention", func() {
		var (
			testTweets []*xscraper.Tweet
			testReq    *service.CreateMentionRequest
		)

		BeforeEach(func() {
			// Setup test tweets
			testTweets = []*xscraper.Tweet{
				{
					RestID: "thread_tweet_1",
					Text:   "This is the first tweet in thread",
					Author: &xscraper.User{
						RestID:          "author1",
						Name:            "Test Author",
						ScreenName:      "testauthor",
						ProfileImageURL: "https://example.com/avatar.jpg",
					},
					CreatedAt: time.Now().Add(-2 * time.Hour),
				},
				{
					RestID: "thread_tweet_2",
					Text:   "This is the second tweet in thread",
					Author: &xscraper.User{
						RestID:          "author1",
						Name:            "Test Author",
						ScreenName:      "testauthor",
						ProfileImageURL: "https://example.com/avatar.jpg",
					},
					CreatedAt: time.Now().Add(-1 * time.Hour),
				},
				{
					RestID: "mention_tweet",
					Text:   "This mentions the thread",
					Author: &xscraper.User{
						RestID:          "mentioner",
						Name:            "Mentioner",
						ScreenName:      "mentioner",
						ProfileImageURL: "https://example.com/mentioner.jpg",
					},
					CreatedAt: time.Now(),
				},
			}
			testReq = &service.CreateMentionRequest{Tweets: testTweets}
		})

		Context("when creating a new mention with new thread", func() {
			It("should create thread and mention successfully", func() {
				result, err := mentionService.CreateMention(ctx, testReq)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.ID).To(Equal("mention_tweet"))
				Expect(result.ThreadID).To(Equal("thread_tweet_2"))
				Expect(result.ThreadAuthor).ToNot(BeNil())
				Expect(result.ThreadAuthor.ID).To(Equal("author1"))
				Expect(result.ThreadAuthor.Name).To(Equal("Test Author"))

				// Verify thread was created with AI summary
				thread, err := threadRepo.GetThreadByID(ctx, "thread_tweet_2")
				Expect(err).ToNot(HaveOccurred())
				Expect(thread.Summary).To(Equal("Mock AI summary for testing"))
				Expect(thread.NumTweets).To(Equal(2))
				Expect(thread.CID).To(Equal("bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u"))

				// Verify mention was created
				mention, err := mentionRepo.GetMentionByID(ctx, "mention_tweet")
				Expect(err).ToNot(HaveOccurred())
				Expect(mention.UserID).To(Equal("mentioner"))
				Expect(mention.ThreadID).To(Equal("thread_tweet_2"))
				Expect(mention.ThreadAuthorID).To(Equal("author1"))
				Expect(mention.ThreadAuthorName).To(Equal("Test Author"))
			})
		})

		Context("when creating mention for existing thread", func() {
			BeforeEach(func() {
				// Pre-create a thread
				existingThread := &model.Thread{
					ID:        "thread_tweet_2",
					Summary:   "Existing thread summary",
					CID:       "QmExisting123",
					NumTweets: 2,
				}
				err := threadRepo.CreateThread(ctx, existingThread)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should create mention without creating new thread", func() {
				result, err := mentionService.CreateMention(ctx, testReq)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.ID).To(Equal("mention_tweet"))
				Expect(result.ThreadID).To(Equal("thread_tweet_2"))

				// Verify thread wasn't modified
				thread, err := threadRepo.GetThreadByID(ctx, "thread_tweet_2")
				Expect(err).ToNot(HaveOccurred())
				Expect(thread.Summary).To(Equal("Existing thread summary"))
				Expect(thread.CID).To(Equal("QmExisting123"))
			})
		})

		Context("when mention already exists", func() {
			BeforeEach(func() {
				// Pre-create thread and mention
				existingThread := &model.Thread{
					ID:        "thread_tweet_2",
					Summary:   "Existing thread summary",
					CID:       "QmExisting123",
					NumTweets: 2,
				}
				err := threadRepo.CreateThread(ctx, existingThread)
				Expect(err).ToNot(HaveOccurred())

				existingMention := &model.Mention{
					ID:                          "mention_tweet",
					UserID:                      "mentioner",
					ThreadID:                    "thread_tweet_2",
					ThreadAuthorID:              "author1",
					ThreadAuthorName:            "Test Author",
					ThreadAuthorScreenName:      "testauthor",
					ThreadAuthorProfileImageURL: "https://example.com/avatar.jpg",
					MentionCreateAt:             time.Now(),
				}
				err = mentionRepo.CreateMention(ctx, existingMention)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return existing mention without duplicating", func() {
				result, err := mentionService.CreateMention(ctx, testReq)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.ID).To(Equal("mention_tweet"))
				Expect(result.ThreadID).To(Equal("thread_tweet_2"))
			})
		})

		Context("when insufficient tweets provided", func() {
			BeforeEach(func() {
				testReq.Tweets = []*xscraper.Tweet{testTweets[0]} // Only one tweet
			})

			It("should return error", func() {
				result, err := mentionService.CreateMention(ctx, testReq)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("no tweets provided"))
			})
		})
	})

	Describe("GetMentionByID", func() {
		var testMention *model.Mention

		BeforeEach(func() {
			// Pre-create a mention
			testMention = &model.Mention{
				ID:                          "test_mention_1",
				UserID:                      "user1",
				ThreadID:                    "thread1",
				ThreadAuthorID:              "author1",
				ThreadAuthorName:            "Test Author",
				ThreadAuthorScreenName:      "testauthor",
				ThreadAuthorProfileImageURL: "https://example.com/avatar.jpg",
				MentionCreateAt:             time.Now(),
			}
			err := mentionRepo.CreateMention(ctx, testMention)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when mention exists", func() {
			It("should return mention summary", func() {
				result, err := mentionService.GetMentionByID(ctx, "test_mention_1")

				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.ID).To(Equal("test_mention_1"))
				Expect(result.ThreadID).To(Equal("thread1"))
				Expect(result.ThreadAuthor.ID).To(Equal("author1"))
				Expect(result.ThreadAuthor.Name).To(Equal("Test Author"))
				Expect(result.ThreadAuthor.ScreenName).To(Equal("testauthor"))
			})
		})

		Context("when mention does not exist", func() {
			It("should return not found error", func() {
				result, err := mentionService.GetMentionByID(ctx, "nonexistent")

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(errors.Is(err, errutil.ErrNotFound)).To(BeTrue())
			})
		})
	})

	Describe("GetMentions", func() {
		BeforeEach(func() {
			// Pre-create test threads
			testThreads := []*model.Thread{
				{
					ID:        "thread1",
					Summary:   "Test thread summary",
					CID:       "QmTest123",
					NumTweets: 3,
				},
				{
					ID:        "thread2",
					Summary:   "Second thread summary",
					CID:       "QmTest456",
					NumTweets: 2,
				},
			}
			for _, thread := range testThreads {
				err := threadRepo.CreateThread(ctx, thread)
				Expect(err).ToNot(HaveOccurred())
			}

			testMentions := []*model.Mention{
				{
					ID:                          "mention1",
					UserID:                      "user1",
					ThreadID:                    "thread1",
					ThreadAuthorID:              "author1",
					ThreadAuthorName:            "Author One",
					ThreadAuthorScreenName:      "author1",
					ThreadAuthorProfileImageURL: "https://example.com/avatar1.jpg",
					MentionCreateAt:             time.Now().Add(-2 * time.Hour),
				},
				{
					ID:                          "mention2",
					UserID:                      "user1",
					ThreadID:                    "thread2",
					ThreadAuthorID:              "author2",
					ThreadAuthorName:            "Author Two",
					ThreadAuthorScreenName:      "author2",
					ThreadAuthorProfileImageURL: "https://example.com/avatar2.jpg",
					MentionCreateAt:             time.Now().Add(-1 * time.Hour),
				},
			}

			for _, mention := range testMentions {
				err := mentionRepo.CreateMention(ctx, mention)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		Context("when user has mentions", func() {
			It("should return mentions with thread details", func() {
				results, total, err := mentionService.GetMentions(ctx, "user1", 10, 0)

				Expect(err).ToNot(HaveOccurred())
				Expect(total).To(Equal(int64(2)))
				Expect(results).To(HaveLen(2))

				// Check mentions contain thread details
				threadIDsSeen := make(map[string]bool)
				for _, result := range results {
					threadIDsSeen[result.ThreadID] = true
					// Each mention should have valid thread details
					Expect(result.ThreadID).To(Or(Equal("thread1"), Equal("thread2")))
					switch result.ThreadID {
					case "thread1":
						Expect(result.ContentPreview).To(Equal("Test thread summary"))
						Expect(result.NumTweets).To(Equal(3))
						Expect(result.CID).To(Equal("QmTest123"))
					case "thread2":
						Expect(result.ContentPreview).To(Equal("Second thread summary"))
						Expect(result.NumTweets).To(Equal(2))
						Expect(result.CID).To(Equal("QmTest456"))
					}
				}
				// Should have mentions from both threads
				Expect(threadIDsSeen).To(HaveKey("thread1"))
				Expect(threadIDsSeen).To(HaveKey("thread2"))
			})
		})

		Context("when user has no mentions", func() {
			It("should return empty list", func() {
				results, total, err := mentionService.GetMentions(ctx, "nonexistent_user", 10, 0)

				Expect(err).ToNot(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
				Expect(results).To(HaveLen(0))
			})
		})

		Context("with pagination", func() {
			It("should respect limit and offset", func() {
				results, total, err := mentionService.GetMentions(ctx, "user1", 1, 1)

				Expect(err).ToNot(HaveOccurred())
				Expect(total).To(Equal(int64(2)))
				Expect(results).To(HaveLen(1))
			})
		})
	})
})
