package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ipfs-force-community/threadmirror/internal/sqlc_generated"
	dbsql "github.com/ipfs-force-community/threadmirror/pkg/database/sql"

	"github.com/jackc/pgx/v5"
)

type BotCookieService struct {
	db *dbsql.DB
}

func NewBotCookieService(db *dbsql.DB) *BotCookieService {
	return &BotCookieService{db: db}
}

// GetBotCookieByEmailAndUsername retrieves a bot cookie by email and username
func (s *BotCookieService) GetBotCookieByEmailAndUsername(ctx context.Context, email, username string) (*sqlc_generated.BotCookie, error) {
	cookie, err := s.db.QueriesFromContext(ctx).GetBotCookieByEmailAndUsername(ctx, sqlc_generated.GetBotCookieByEmailAndUsernameParams{
		Email:    email,
		Username: username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get bot cookie: %w", err)
	}
	return &cookie, nil
}

// CreateBotCookie creates a new bot cookie
func (s *BotCookieService) CreateBotCookie(ctx context.Context, email, username string, cookiesData interface{}) (*sqlc_generated.BotCookie, error) {
	// Marshal cookies data to JSON
	jsonData, err := json.Marshal(cookiesData)
	if err != nil {
		return nil, fmt.Errorf("marshal cookies data: %w", err)
	}

	cookie, err := s.db.QueriesFromContext(ctx).CreateBotCookie(ctx, sqlc_generated.CreateBotCookieParams{
		Email:       email,
		Username:    username,
		CookiesData: jsonData,
	})
	if err != nil {
		return nil, fmt.Errorf("create bot cookie: %w", err)
	}
	return &cookie, nil
}

// UpdateBotCookie updates an existing bot cookie
func (s *BotCookieService) UpdateBotCookie(ctx context.Context, id int32, email, username string, cookiesData interface{}) error {
	// Marshal cookies data to JSON
	jsonData, err := json.Marshal(cookiesData)
	if err != nil {
		return fmt.Errorf("marshal cookies data: %w", err)
	}

	err = s.db.QueriesFromContext(ctx).UpdateBotCookie(ctx, sqlc_generated.UpdateBotCookieParams{
		ID:          id,
		Email:       email,
		Username:    username,
		CookiesData: jsonData,
	})
	if err != nil {
		return fmt.Errorf("update bot cookie: %w", err)
	}
	return nil
}

// SoftDeleteBotCookie soft deletes a bot cookie
func (s *BotCookieService) SoftDeleteBotCookie(ctx context.Context, id int32) error {
	err := s.db.QueriesFromContext(ctx).SoftDeleteBotCookie(ctx, sqlc_generated.SoftDeleteBotCookieParams{ID: id})
	if err != nil {
		return fmt.Errorf("soft delete bot cookie: %w", err)
	}
	return nil
}

// ListBotCookies lists bot cookies with pagination
func (s *BotCookieService) ListBotCookies(ctx context.Context, limit, offset int) ([]sqlc_generated.BotCookie, int64, error) {
	cookies, err := s.db.QueriesFromContext(ctx).ListBotCookies(ctx, sqlc_generated.ListBotCookiesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list bot cookies: %w", err)
	}

	count, err := s.db.QueriesFromContext(ctx).CountBotCookies(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count bot cookies: %w", err)
	}

	return cookies, count, nil
}

// LoadCookies loads cookies for a specific email and username
func (s *BotCookieService) LoadCookies(ctx context.Context, email, username string) ([]*http.Cookie, error) {
	cookie, err := s.GetBotCookieByEmailAndUsername(ctx, email, username)
	if err != nil {
		return nil, fmt.Errorf("get bot cookie: %w", err)
	}

	var cookies []*http.Cookie
	err = json.Unmarshal(cookie.CookiesData, &cookies)
	if err != nil {
		return nil, fmt.Errorf("unmarshal cookies: %w", err)
	}

	return cookies, nil
}

// SaveCookies saves cookies for a specific email and username
func (s *BotCookieService) SaveCookies(ctx context.Context, email, username string, cookies []*http.Cookie) error {
	// Try to get existing cookie first
	existingCookie, err := s.GetBotCookieByEmailAndUsername(ctx, email, username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Create new cookie
			_, err = s.CreateBotCookie(ctx, email, username, cookies)
			return err
		}
		return fmt.Errorf("get existing cookie: %w", err)
	}

	// Update existing cookie
	return s.UpdateBotCookie(ctx, existingCookie.ID, email, username, cookies)
}
