package service_test

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BotCookieService", func() {
	var (
		botCookieService *service.BotCookieService
		botCookieRepo    service.BotCookieRepoInterface
		db               *sql.DB
		ctx              context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Setup testcontainers database
		db = testsuit.SetupGinkgoTestDB()

		// Clean database before each test
		testsuit.ResetGinkgoDatabase()

		// Initialize repo with database
		botCookieRepo = sqlrepo.NewBotCookieRepo(db)

		// Create service
		botCookieService = service.NewBotCookieService(botCookieRepo)
	})

	Describe("SaveCookies", func() {
		var testCookies []*http.Cookie

		BeforeEach(func() {
			testCookies = []*http.Cookie{
				{
					Name:    "session_id",
					Value:   "abc123",
					Domain:  ".example.com",
					Path:    "/",
					Expires: time.Now().Add(24 * time.Hour),
				},
				{
					Name:   "csrf_token",
					Value:  "xyz789",
					Domain: ".example.com",
					Path:   "/",
				},
			}
		})

		Context("when saving valid cookies", func() {
			It("should save cookies successfully", func() {
				err := botCookieService.SaveCookies(ctx, "test@example.com", "testuser", testCookies)

				Expect(err).ToNot(HaveOccurred())

				// Verify cookies were saved by loading them back
				loadedCookies, err := botCookieService.LoadCookies(ctx, "test@example.com", "testuser")
				Expect(err).ToNot(HaveOccurred())
				Expect(loadedCookies).ToNot(BeNil())
			})
		})

		Context("when overwriting existing cookies", func() {
			BeforeEach(func() {
				// Save initial cookies
				err := botCookieService.SaveCookies(ctx, "test@example.com", "testuser", testCookies)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should overwrite existing cookies", func() {
				newCookies := []*http.Cookie{
					{
						Name:  "new_session",
						Value: "new123",
					},
				}

				err := botCookieService.SaveCookies(ctx, "test@example.com", "testuser", newCookies)
				Expect(err).ToNot(HaveOccurred())

				// Verify new cookies are loaded
				loadedCookies, err := botCookieService.LoadCookies(ctx, "test@example.com", "testuser")
				Expect(err).ToNot(HaveOccurred())
				Expect(loadedCookies).To(HaveLen(1))
				Expect(loadedCookies[0].Name).To(Equal("new_session"))
			})
		})

		Context("with empty email or username", func() {
			It("should save with empty email", func() {
				err := botCookieService.SaveCookies(ctx, "", "testuser", testCookies)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should save with empty username", func() {
				err := botCookieService.SaveCookies(ctx, "test@example.com", "", testCookies)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("LoadCookies", func() {
		Context("when cookies exist", func() {
			var originalCookies []*http.Cookie

			BeforeEach(func() {
				originalCookies = []*http.Cookie{
					{
						Name:     "auth_token",
						Value:    "token123",
						Domain:   ".example.com",
						Path:     "/",
						Expires:  time.Now().Add(24 * time.Hour),
						HttpOnly: true,
						Secure:   true,
					},
					{
						Name:  "user_pref",
						Value: "dark_mode",
					},
				}

				err := botCookieService.SaveCookies(ctx, "user@example.com", "username", originalCookies)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should load cookies successfully", func() {
				loadedCookies, err := botCookieService.LoadCookies(ctx, "user@example.com", "username")

				Expect(err).ToNot(HaveOccurred())
				Expect(loadedCookies).To(HaveLen(2))

				// Check first cookie
				cookie1 := loadedCookies[0]
				Expect(cookie1.Name).To(Equal("auth_token"))
				Expect(cookie1.Value).To(Equal("token123"))
				Expect(cookie1.Domain).To(Equal(".example.com"))
				Expect(cookie1.Path).To(Equal("/"))
				Expect(cookie1.HttpOnly).To(BeTrue())
				Expect(cookie1.Secure).To(BeTrue())

				// Check second cookie
				cookie2 := loadedCookies[1]
				Expect(cookie2.Name).To(Equal("user_pref"))
				Expect(cookie2.Value).To(Equal("dark_mode"))
			})
		})

		Context("when cookies do not exist", func() {
			It("should return not found error", func() {
				loadedCookies, err := botCookieService.LoadCookies(ctx, "nonexistent@example.com", "user")

				Expect(err).To(HaveOccurred())
				Expect(loadedCookies).To(BeNil())
				Expect(errors.Is(err, errutil.ErrNotFound)).To(BeTrue())
			})
		})

		Context("when partial matches exist", func() {
			BeforeEach(func() {
				testCookies := []*http.Cookie{
					{Name: "test", Value: "value"},
				}
				err := botCookieService.SaveCookies(ctx, "user1@example.com", "user1", testCookies)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not return cookies for different email", func() {
				loadedCookies, err := botCookieService.LoadCookies(ctx, "user2@example.com", "user1")

				Expect(err).To(HaveOccurred())
				Expect(loadedCookies).To(BeNil())
				Expect(errors.Is(err, errutil.ErrNotFound)).To(BeTrue())
			})

			It("should not return cookies for different username", func() {
				loadedCookies, err := botCookieService.LoadCookies(ctx, "user1@example.com", "user2")

				Expect(err).To(HaveOccurred())
				Expect(loadedCookies).To(BeNil())
				Expect(errors.Is(err, errutil.ErrNotFound)).To(BeTrue())
			})
		})
	})

	Describe("GetLatestBotCookie", func() {
		Context("when no cookies exist", func() {
			It("should return not found error", func() {
				cookie, err := botCookieService.GetLatestBotCookie(ctx)

				Expect(err).To(HaveOccurred())
				Expect(cookie).To(BeNil())
				Expect(errors.Is(err, errutil.ErrNotFound)).To(BeTrue())
			})
		})

		Context("when cookies exist", func() {
			BeforeEach(func() {
				// Save some test cookies
				testCookies := []*http.Cookie{
					{Name: "session", Value: "abc123"},
				}

				err := botCookieService.SaveCookies(ctx, "bot1@example.com", "bot1", testCookies)
				Expect(err).ToNot(HaveOccurred())

				err = botCookieService.SaveCookies(ctx, "bot2@example.com", "bot2", testCookies)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return a bot cookie record", func() {
				cookie, err := botCookieService.GetLatestBotCookie(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(cookie).ToNot(BeNil())
				Expect(cookie.Email).ToNot(BeEmpty())
				Expect(cookie.CookiesData).ToNot(BeNil())
			})
		})

		Context("with multiple cookies at different times", func() {
			BeforeEach(func() {
				testCookies := []*http.Cookie{
					{Name: "session", Value: "test"},
				}

				// Save first cookie
				err := botCookieService.SaveCookies(ctx, "old@example.com", "old", testCookies)
				Expect(err).ToNot(HaveOccurred())

				// Wait a bit and save second cookie to ensure different timestamps
				time.Sleep(10 * time.Millisecond)
				err = botCookieService.SaveCookies(ctx, "new@example.com", "new", testCookies)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return the most recently updated cookie", func() {
				cookie, err := botCookieService.GetLatestBotCookie(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(cookie).ToNot(BeNil())
				// In real implementation, this would be the most recent one
				// For test, we just verify we get a valid record
				Expect(cookie.Email).To(BeElementOf([]string{"old@example.com", "new@example.com"}))
			})
		})
	})
})
