package service_test

import (
	"context"

	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessedMarkService", func() {
	var (
		processedMarkService *service.ProcessedMarkService
		processedMarkRepo    service.ProcessedMarkRepoInterface
		db                   *sql.DB
		ctx                  context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		db = testsuit.SetupGinkgoTestDB()

		// Clean database before each test
		testsuit.ResetGinkgoDatabase()

		// Initialize repo with database
		processedMarkRepo = sqlrepo.NewProcessedMarkRepo(db)

		// Create service
		processedMarkService = service.NewProcessedMarkService(processedMarkRepo)
	})

	Describe("IsProcessed", func() {
		Context("when mark does not exist", func() {
			It("should return false", func() {
				processed, err := processedMarkService.IsProcessed(ctx, "nonexistent_key", "test_type")

				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeFalse())
			})
		})

		Context("when mark exists", func() {
			BeforeEach(func() {
				// Pre-mark something as processed
				err := processedMarkService.MarkProcessed(ctx, "existing_key", "test_type")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return true for existing mark", func() {
				processed, err := processedMarkService.IsProcessed(ctx, "existing_key", "test_type")

				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())
			})

			It("should return false for different key", func() {
				processed, err := processedMarkService.IsProcessed(ctx, "different_key", "test_type")

				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeFalse())
			})

			It("should return false for different type", func() {
				processed, err := processedMarkService.IsProcessed(ctx, "existing_key", "different_type")

				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeFalse())
			})
		})

		Context("with multiple marks", func() {
			BeforeEach(func() {
				// Mark multiple items as processed
				testMarks := []struct {
					key string
					typ string
				}{
					{"key1", "type_a"},
					{"key2", "type_a"},
					{"key1", "type_b"},
					{"key3", "type_c"},
				}

				for _, mark := range testMarks {
					err := processedMarkService.MarkProcessed(ctx, mark.key, mark.typ)
					Expect(err).ToNot(HaveOccurred())
				}
			})

			It("should correctly identify each mark", func() {
				// Test existing marks
				processed, err := processedMarkService.IsProcessed(ctx, "key1", "type_a")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())

				processed, err = processedMarkService.IsProcessed(ctx, "key2", "type_a")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())

				processed, err = processedMarkService.IsProcessed(ctx, "key1", "type_b")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())

				processed, err = processedMarkService.IsProcessed(ctx, "key3", "type_c")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())

				// Test non-existing marks
				processed, err = processedMarkService.IsProcessed(ctx, "key2", "type_b")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeFalse())

				processed, err = processedMarkService.IsProcessed(ctx, "key4", "type_a")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeFalse())
			})
		})
	})

	Describe("MarkProcessed", func() {
		Context("when marking a new item", func() {
			It("should mark successfully and be queryable", func() {
				err := processedMarkService.MarkProcessed(ctx, "new_key", "new_type")

				Expect(err).ToNot(HaveOccurred())

				// Verify it's marked as processed
				processed, err := processedMarkService.IsProcessed(ctx, "new_key", "new_type")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())
			})
		})

		Context("when marking an already processed item", func() {
			BeforeEach(func() {
				// Pre-mark an item
				err := processedMarkService.MarkProcessed(ctx, "duplicate_key", "duplicate_type")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return ErrMentionAlreadyProcessed", func() {
				err := processedMarkService.MarkProcessed(ctx, "duplicate_key", "duplicate_type")

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(service.ErrMentionAlreadyProcessed))
			})
		})

		Context("with empty key or type", func() {
			It("should handle empty key", func() {
				err := processedMarkService.MarkProcessed(ctx, "", "test_type")

				Expect(err).ToNot(HaveOccurred())

				// Verify it can be queried
				processed, err := processedMarkService.IsProcessed(ctx, "", "test_type")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())
			})

			It("should handle empty type", func() {
				err := processedMarkService.MarkProcessed(ctx, "test_key", "")

				Expect(err).ToNot(HaveOccurred())

				// Verify it can be queried
				processed, err := processedMarkService.IsProcessed(ctx, "test_key", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())
			})

			It("should handle both empty", func() {
				err := processedMarkService.MarkProcessed(ctx, "", "")

				Expect(err).ToNot(HaveOccurred())

				// Verify it can be queried
				processed, err := processedMarkService.IsProcessed(ctx, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())
			})
		})

		Context("with special characters", func() {
			It("should handle keys with special characters", func() {
				specialKeys := []string{
					"key-with-dashes",
					"key_with_underscores",
					"key.with.dots",
					"key with spaces",
					"key/with/slashes",
					"key@with@symbols",
					"key#with#hash",
					"ключ_на_русском", // Russian characters
					"中文键",             // Chinese characters
				}

				for _, key := range specialKeys {
					err := processedMarkService.MarkProcessed(ctx, key, "special_type")
					Expect(err).ToNot(HaveOccurred())

					// Verify it can be queried
					processed, err := processedMarkService.IsProcessed(ctx, key, "special_type")
					Expect(err).ToNot(HaveOccurred())
					Expect(processed).To(BeTrue())
				}
			})

			It("should handle types with special characters", func() {
				specialTypes := []string{
					"type-with-dashes",
					"type_with_underscores",
					"type.with.dots",
					"type with spaces",
					"тип_на_русском", // Russian
					"中文类型",           // Chinese
				}

				for _, typ := range specialTypes {
					err := processedMarkService.MarkProcessed(ctx, "test_key", typ)
					Expect(err).ToNot(HaveOccurred())

					// Verify it can be queried
					processed, err := processedMarkService.IsProcessed(ctx, "test_key", typ)
					Expect(err).ToNot(HaveOccurred())
					Expect(processed).To(BeTrue())
				}
			})
		})
	})

	Describe("Integration scenarios", func() {
		Context("with typical usage patterns", func() {
			It("should handle mention processing workflow", func() {
				mentionID := "mention_12345"
				mentionType := "twitter_mention"

				// Check if not processed initially
				processed, err := processedMarkService.IsProcessed(ctx, mentionID, mentionType)
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeFalse())

				// Mark as processed
				err = processedMarkService.MarkProcessed(ctx, mentionID, mentionType)
				Expect(err).ToNot(HaveOccurred())

				// Verify it's now processed
				processed, err = processedMarkService.IsProcessed(ctx, mentionID, mentionType)
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeTrue())

				// Try to mark again (should fail)
				err = processedMarkService.MarkProcessed(ctx, mentionID, mentionType)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(service.ErrMentionAlreadyProcessed))
			})

			It("should handle batch processing simulation", func() {
				batchItems := []struct {
					id  string
					typ string
				}{
					{"item_001", "batch_type"},
					{"item_002", "batch_type"},
					{"item_003", "batch_type"},
					{"item_004", "batch_type"},
					{"item_005", "batch_type"},
				}

				// Mark all items as processed
				for _, item := range batchItems {
					err := processedMarkService.MarkProcessed(ctx, item.id, item.typ)
					Expect(err).ToNot(HaveOccurred())
				}

				// Verify all items are marked
				for _, item := range batchItems {
					processed, err := processedMarkService.IsProcessed(ctx, item.id, item.typ)
					Expect(err).ToNot(HaveOccurred())
					Expect(processed).To(BeTrue())
				}

				// Check that unmarked items are still not processed
				processed, err := processedMarkService.IsProcessed(ctx, "item_006", "batch_type")
				Expect(err).ToNot(HaveOccurred())
				Expect(processed).To(BeFalse())
			})
		})
	})
})
