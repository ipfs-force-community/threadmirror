package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

var _ = Describe("BotCookieService", func() {
	var (
		botCookieService *service.BotCookieService
		db               *sql.DB
		ctx              context.Context
		suite            *testsuit.ContainerTestSuite
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		suite = testsuit.SetupContainerTestSuite(&testing.T{})
		db = suite.DB

		// Create service
		botCookieService = service.NewBotCookieService(db)

		// Reset database for clean test state
		suite.ResetDatabase(&testing.T{})
	})

	AfterEach(func() {
		if suite != nil {
			suite.TearDown(&testing.T{})
		}
	})

	Describe("CreateBotCookie", func() {
		It("should create a new bot cookie successfully", func() {
			email := "test@example.com"
			username := "testuser"
			cookies := []*http.Cookie{
				{Name: "session", Value: "abc123", Domain: ".twitter.com"},
				{Name: "csrf", Value: "xyz789", Domain: ".twitter.com"},
			}

			result, err := botCookieService.CreateBotCookie(ctx, email, username, cookies)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Email).To(Equal(email))
			Expect(result.Username).To(Equal(username))

			// Verify cookies were stored correctly
			var storedCookies []*http.Cookie
			err = json.Unmarshal(result.CookiesData, &storedCookies)
			Expect(err).ToNot(HaveOccurred())
			Expect(storedCookies).To(HaveLen(2))
		})

		It("should return error for duplicate email and username", func() {
			email := "test@example.com"
			username := "testuser"
			cookies := []*http.Cookie{
				{Name: "session", Value: "abc123"},
			}

			// Create first cookie
			_, err := botCookieService.CreateBotCookie(ctx, email, username, cookies)
			Expect(err).ToNot(HaveOccurred())

			// Try to create duplicate
			_, err = botCookieService.CreateBotCookie(ctx, email, username, cookies)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetBotCookieByEmailAndUsername", func() {
		It("should retrieve bot cookie by email and username", func() {
			email := "test@example.com"
			username := "testuser"
			cookies := []*http.Cookie{
				{Name: "session", Value: "abc123"},
			}

			// Create cookie first
			created, err := botCookieService.CreateBotCookie(ctx, email, username, cookies)
			Expect(err).ToNot(HaveOccurred())

			// Retrieve it
			retrieved, err := botCookieService.GetBotCookieByEmailAndUsername(ctx, email, username)
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved.ID).To(Equal(created.ID))
			Expect(retrieved.Email).To(Equal(email))
			Expect(retrieved.Username).To(Equal(username))
		})

		It("should return error for non-existent cookie", func() {
			_, err := botCookieService.GetBotCookieByEmailAndUsername(ctx, "nonexistent@example.com", "nonuser")
			Expect(err).To(Equal(service.ErrBotCookieNotFound))
		})
	})

	Describe("UpdateBotCookie", func() {
		It("should update existing bot cookie", func() {
			email := "test@example.com"
			username := "testuser"
			cookies := []*http.Cookie{
				{Name: "session", Value: "abc123"},
			}

			// Create cookie first
			created, err := botCookieService.CreateBotCookie(ctx, email, username, cookies)
			Expect(err).ToNot(HaveOccurred())

			// Update with new cookies
			newCookies := []*http.Cookie{
				{Name: "session", Value: "def456"},
				{Name: "csrf", Value: "ghi789"},
			}

			err = botCookieService.UpdateBotCookie(ctx, created.ID, email, username, newCookies)
			Expect(err).ToNot(HaveOccurred())

			// Verify update
			updated, err := botCookieService.GetBotCookieByEmailAndUsername(ctx, email, username)
			Expect(err).ToNot(HaveOccurred())

			var storedCookies []*http.Cookie
			err = json.Unmarshal(updated.CookiesData, &storedCookies)
			Expect(err).ToNot(HaveOccurred())
			Expect(storedCookies).To(HaveLen(2))
			Expect(storedCookies[0].Value).To(Equal("def456"))
		})
	})

	Describe("ListBotCookies", func() {
		It("should list bot cookies with pagination", func() {
			// Create multiple cookies
			for i := 0; i < 5; i++ {
				email := fmt.Sprintf("test%d@example.com", i)
				username := fmt.Sprintf("user%d", i)
				cookies := []*http.Cookie{
					{Name: "session", Value: fmt.Sprintf("value%d", i)},
				}
				_, err := botCookieService.CreateBotCookie(ctx, email, username, cookies)
				Expect(err).ToNot(HaveOccurred())
			}

			// List with pagination
			cookies, total, err := botCookieService.ListBotCookies(ctx, 3, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(cookies).To(HaveLen(3))
			Expect(total).To(Equal(int64(5)))

			// Test second page
			cookies, total, err = botCookieService.ListBotCookies(ctx, 3, 3)
			Expect(err).ToNot(HaveOccurred())
			Expect(cookies).To(HaveLen(2))
			Expect(total).To(Equal(int64(5)))
		})
	})

	Describe("SoftDeleteBotCookie", func() {
		It("should soft delete bot cookie", func() {
			email := "test@example.com"
			username := "testuser"
			cookies := []*http.Cookie{
				{Name: "session", Value: "abc123"},
			}

			// Create cookie first
			created, err := botCookieService.CreateBotCookie(ctx, email, username, cookies)
			Expect(err).ToNot(HaveOccurred())

			// Soft delete
			err = botCookieService.SoftDeleteBotCookie(ctx, created.ID)
			Expect(err).ToNot(HaveOccurred())

			// Verify it's not accessible
			_, err = botCookieService.GetBotCookieByEmailAndUsername(ctx, email, username)
			Expect(err).To(Equal(service.ErrBotCookieNotFound))
		})
	})

	Describe("LoadCookies and SaveCookies", func() {
		It("should load and save cookies for user", func() {
			email := "test@example.com"
			username := "testuser"
			cookies := []*http.Cookie{
				{Name: "session", Value: "abc123"},
				{Name: "csrf", Value: "xyz789"},
			}

			// Save cookies
			err := botCookieService.SaveCookies(ctx, email, username, cookies)
			Expect(err).ToNot(HaveOccurred())

			// Load cookies
			loadedCookies, err := botCookieService.LoadCookies(ctx, email, username)
			Expect(err).ToNot(HaveOccurred())
			Expect(loadedCookies).To(HaveLen(2))
			Expect(loadedCookies[0].Name).To(Equal("session"))
			Expect(loadedCookies[0].Value).To(Equal("abc123"))
		})

		It("should update existing cookies when saving", func() {
			email := "test@example.com"
			username := "testuser"

			// Save initial cookies
			initialCookies := []*http.Cookie{
				{Name: "session", Value: "abc123"},
			}
			err := botCookieService.SaveCookies(ctx, email, username, initialCookies)
			Expect(err).ToNot(HaveOccurred())

			// Save updated cookies
			updatedCookies := []*http.Cookie{
				{Name: "session", Value: "def456"},
				{Name: "csrf", Value: "ghi789"},
			}
			err = botCookieService.SaveCookies(ctx, email, username, updatedCookies)
			Expect(err).ToNot(HaveOccurred())

			// Load and verify
			loadedCookies, err := botCookieService.LoadCookies(ctx, email, username)
			Expect(err).ToNot(HaveOccurred())
			Expect(loadedCookies).To(HaveLen(2))
			Expect(loadedCookies[0].Value).To(Equal("def456"))
		})
	})
})
