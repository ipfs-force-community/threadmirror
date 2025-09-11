package service_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

var _ = Describe("ProcessedMarkService", func() {
	var (
		processedMarkService *service.ProcessedMarkService
		db                   *sql.DB
		ctx                  context.Context
		suite                *testsuit.ContainerTestSuite
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		suite = testsuit.SetupContainerTestSuite(&testing.T{})
		db = suite.DB

		// Create service
		processedMarkService = service.NewProcessedMarkService(db)

		// Reset database for clean test state
		suite.ResetDatabase(&testing.T{})
	})

	AfterEach(func() {
		if suite != nil {
			suite.TearDown(&testing.T{})
		}
	})

	Describe("MarkAsProcessed", func() {
		It("should mark a resource as processed successfully", func() {
			resourceID := "test-resource-123"
			resourceType := "test-type"

			err := processedMarkService.MarkAsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())

			// Verify it's marked as processed
			isProcessed, err := processedMarkService.IsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())
			Expect(isProcessed).To(BeTrue())
		})

		It("should handle duplicate marks gracefully", func() {
			resourceID := "test-resource-123"
			resourceType := "test-type"

			// Mark first time
			err := processedMarkService.MarkAsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())

			// Mark second time - should not error
			err = processedMarkService.MarkAsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("IsProcessed", func() {
		It("should return false for unprocessed resource", func() {
			isProcessed, err := processedMarkService.IsProcessed(ctx, "nonexistent", "test-type")
			Expect(err).ToNot(HaveOccurred())
			Expect(isProcessed).To(BeFalse())
		})

		It("should return true for processed resource", func() {
			resourceID := "test-resource-123"
			resourceType := "test-type"

			// Mark as processed
			err := processedMarkService.MarkAsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())

			// Check if processed
			isProcessed, err := processedMarkService.IsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())
			Expect(isProcessed).To(BeTrue())
		})
	})

	Describe("CleanupOldMarks", func() {
		It("should cleanup old processed marks", func() {
			resourceID := "test-resource-123"
			resourceType := "test-type"

			// Mark as processed
			err := processedMarkService.MarkAsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())

			// Cleanup marks older than 1 second (should not affect recent mark)
			err = processedMarkService.CleanupOldMarks(ctx, time.Second)
			Expect(err).ToNot(HaveOccurred())

			// Verify mark still exists
			isProcessed, err := processedMarkService.IsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())
			Expect(isProcessed).To(BeTrue())

			// Cleanup marks older than 0 (should remove all marks)
			err = processedMarkService.CleanupOldMarks(ctx, 0)
			Expect(err).ToNot(HaveOccurred())

			// Verify mark is removed
			isProcessed, err = processedMarkService.IsProcessed(ctx, resourceID, resourceType)
			Expect(err).ToNot(HaveOccurred())
			Expect(isProcessed).To(BeFalse())
		})
	})
})
