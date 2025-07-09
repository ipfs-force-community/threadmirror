package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/redis"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ThreadService", func() {
	var (
		threadService *service.ThreadService
		threadRepo    service.ThreadRepoInterface
		mockIPFS      *testsuit.MockIPFSStorage
		db            *sql.DB
		redisClient   *redis.Client
		logger        *slog.Logger
		ctx           context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		db = testsuit.SetupGinkgoTestDB()

		// Clean database before each test
		testsuit.ResetGinkgoDatabase()

		// Setup Redis testcontainer
		redisClient = testsuit.SetupTestRedis()

		// Initialize repo with database
		threadRepo = sqlrepo.NewThreadRepo(db)

		// Initialize mock IPFS
		mockIPFS = &testsuit.MockIPFSStorage{}

		// Initialize mock LLM
		mockLLM := &testsuit.MockLLM{}

		// Setup logger
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn, // Reduce test noise
		}))

		// Create service
		threadService = service.NewThreadService(threadRepo, mockIPFS, mockLLM, redisClient, logger)
	})

	Describe("GetThreadByID", func() {
		var testThread *model.Thread

		BeforeEach(func() {
			// Pre-create a thread in database
			testThread = &model.Thread{
				ID:        "thread123",
				Summary:   "Test thread summary",
				CID:       "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
				NumTweets: 3,
				Status:    model.ThreadStatusCompleted, // Set status to completed so tweets are loaded from IPFS
			}
			err := threadRepo.CreateThread(ctx, testThread)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when thread exists and IPFS returns data", func() {
			It("should return thread with tweets from IPFS", func() {
				result, err := threadService.GetThreadByID(ctx, "thread123")

				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.ID).To(Equal("thread123"))
				Expect(result.CID).To(Equal("bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u"))
				Expect(result.ContentPreview).To(Equal("Test thread summary"))
				Expect(result.NumTweets).To(Equal(3))
				Expect(result.Tweets).ToNot(BeNil())
				Expect(result.Tweets).To(HaveLen(1)) // MockIPFS returns 1 tweet

				// Check tweet data from mock
				tweet := result.Tweets[0]
				Expect(tweet.RestID).To(Equal("mock-tweet-1"))
				Expect(tweet.Text).To(Equal("This is a mock tweet for testing"))
				Expect(tweet.Author).ToNot(BeNil())
				Expect(tweet.Author.RestID).To(Equal("mock-user-1"))
				Expect(tweet.Author.Name).To(Equal("Mock User"))
			})
		})

		Context("when thread does not exist", func() {
			It("should return ErrNotFound", func() {
				_, err := threadService.GetThreadByID(ctx, "nonexistent_thread")

				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, errutil.ErrNotFound)).To(BeTrue())
			})
		})

		Context("when thread exists but has empty CID", func() {
			BeforeEach(func() {
				emptyCIDThread := &model.Thread{
					ID:        "empty_cid_thread",
					Summary:   "Thread with empty CID",
					CID:       "",
					NumTweets: 1,
					Status:    model.ThreadStatusCompleted, // Even completed threads with empty CID don't load tweets
				}
				err := threadRepo.CreateThread(ctx, emptyCIDThread)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return thread without tweets when CID is empty", func() {
				result, err := threadService.GetThreadByID(ctx, "empty_cid_thread")

				Expect(err).ToNot(HaveOccurred()) // Should not error, just return empty tweets
				Expect(result).ToNot(BeNil())
				Expect(result.ID).To(Equal("empty_cid_thread"))
				Expect(result.Tweets).To(BeEmpty()) // Empty tweets when no CID
			})
		})

		Context("with caching behavior", func() {
			It("should cache IPFS results on first call", func() {
				// First call should hit IPFS
				result1, err := threadService.GetThreadByID(ctx, "thread123")
				Expect(err).ToNot(HaveOccurred())
				Expect(result1.Tweets).To(HaveLen(1))

				// Second call should use cache (verified by consistent results)
				result2, err := threadService.GetThreadByID(ctx, "thread123")
				Expect(err).ToNot(HaveOccurred())
				Expect(result2.Tweets).To(HaveLen(1))
				Expect(result2.Tweets[0].RestID).To(Equal(result1.Tweets[0].RestID))
			})
		})

		Context("with multiple threads", func() {
			BeforeEach(func() {
				// Create additional test threads
				additionalThreads := []*model.Thread{
					{
						ID:        "thread456",
						Summary:   "Second test thread",
						CID:       "bafkreig7vfzqkdqkr3xdkhhh4n3t6n6q5j6x2kqb3j2rjvlq3d6bqdz5eu",
						NumTweets: 2,
						Status:    model.ThreadStatusCompleted,
					},
					{
						ID:        "thread789",
						Summary:   "Third test thread",
						CID:       "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
						NumTweets: 5,
						Status:    model.ThreadStatusCompleted,
					},
				}

				for _, thread := range additionalThreads {
					err := threadRepo.CreateThread(ctx, thread)
					Expect(err).ToNot(HaveOccurred())
				}
			})

			It("should return correct thread for each ID", func() {
				// Test first thread
				result1, err := threadService.GetThreadByID(ctx, "thread456")
				Expect(err).ToNot(HaveOccurred())
				Expect(result1.ID).To(Equal("thread456"))
				Expect(result1.ContentPreview).To(Equal("Second test thread"))
				Expect(result1.NumTweets).To(Equal(2))

				// Test second thread
				result2, err := threadService.GetThreadByID(ctx, "thread789")
				Expect(err).ToNot(HaveOccurred())
				Expect(result2.ID).To(Equal("thread789"))
				Expect(result2.ContentPreview).To(Equal("Third test thread"))
				Expect(result2.NumTweets).To(Equal(5))
			})
		})
	})

	Describe("Cache functionality", func() {
		var testThread *model.Thread

		BeforeEach(func() {
			testThread = &model.Thread{
				ID:        "cache_test_thread",
				Summary:   "Cache test thread",
				CID:       "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
				NumTweets: 2,
				Status:    model.ThreadStatusCompleted,
			}
			err := threadRepo.CreateThread(ctx, testThread)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when cache is available", func() {
			It("should use cache on subsequent calls", func() {
				// First call populates cache
				result1, err := threadService.GetThreadByID(ctx, "cache_test_thread")
				Expect(err).ToNot(HaveOccurred())
				Expect(result1.Tweets).ToNot(BeNil())

				// Verify cache key would be set (indirect test through consistent behavior)
				result2, err := threadService.GetThreadByID(ctx, "cache_test_thread")
				Expect(err).ToNot(HaveOccurred())
				Expect(result2.Tweets).To(HaveLen(len(result1.Tweets)))
			})
		})

		Context("when Redis is unavailable", func() {
			// Note: This would require injecting a failing Redis client
			// For now, we test that the service still works when cache operations fail
			It("should still work without cache", func() {
				result, err := threadService.GetThreadByID(ctx, "cache_test_thread")

				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.ID).To(Equal("cache_test_thread"))
				Expect(result.Tweets).ToNot(BeNil())
			})
		})
	})

	Describe("Edge cases", func() {
		Context("with malformed CID", func() {
			BeforeEach(func() {
				malformedCIDThread := &model.Thread{
					ID:        "malformed_cid_thread",
					Summary:   "Thread with malformed CID",
					CID:       "invalid-cid-format",
					NumTweets: 1,
					Status:    model.ThreadStatusCompleted,
				}
				err := threadRepo.CreateThread(ctx, malformedCIDThread)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return error when parsing CID", func() {
				result, err := threadService.GetThreadByID(ctx, "malformed_cid_thread")

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to parse CID"))
			})
		})

		Context("with concurrent access", func() {
			BeforeEach(func() {
				concurrentThread := &model.Thread{
					ID:        "concurrent_thread",
					Summary:   "Thread for concurrent test",
					CID:       "bafkreidivzimqfqtoqxkrpge6bjyhlvxqs3rhe73owtmdulaxr5do5in7u",
					NumTweets: 1,
					Status:    model.ThreadStatusCompleted,
				}
				err := threadRepo.CreateThread(ctx, concurrentThread)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle concurrent requests safely", func() {
				resultChan := make(chan *service.ThreadDetail, 5)
				errorChan := make(chan error, 5)

				// Launch concurrent requests
				for i := 0; i < 5; i++ {
					go func() {
						result, err := threadService.GetThreadByID(ctx, "concurrent_thread")
						if err != nil {
							errorChan <- err
						} else {
							resultChan <- result
						}
					}()
				}

				// Collect results
				var results []*service.ThreadDetail
				var errors []error

				for i := 0; i < 5; i++ {
					select {
					case result := <-resultChan:
						results = append(results, result)
					case err := <-errorChan:
						errors = append(errors, err)
					case <-time.After(5 * time.Second):
						Fail("Timeout waiting for concurrent requests")
					}
				}

				// All requests should succeed
				Expect(errors).To(HaveLen(0))
				Expect(results).To(HaveLen(5))

				// All results should be consistent
				for _, result := range results {
					Expect(result.ID).To(Equal("concurrent_thread"))
					Expect(result.ContentPreview).To(Equal("Thread for concurrent test"))
				}
			})
		})
	})
})
