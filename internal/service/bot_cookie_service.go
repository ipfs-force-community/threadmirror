package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/datatypes"
)

// BotCookieRepoInterface defines the interface for bot cookie repo operations
type BotCookieRepoInterface interface {
	GetCookies(ctx context.Context, email, username string) (datatypes.JSON, error)
	SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error
}

// BotCookieService provides business logic for bot cookie management
type BotCookieService struct {
	repo BotCookieRepoInterface
}

// NewBotCookieService creates a new bot cookie service
func NewBotCookieService(repo BotCookieRepoInterface) *BotCookieService {
	return &BotCookieService{
		repo: repo,
	}
}

// SaveCookies saves HTTP cookies to database
func (s *BotCookieService) SaveCookies(ctx context.Context, email, username string, cookies []*http.Cookie) error {
	err := s.repo.SaveCookies(ctx, email, username, cookies)
	if err != nil {
		return fmt.Errorf("failed to save cookies to database: %w", err)
	}
	return nil
}

// LoadCookies loads HTTP cookies from database
func (s *BotCookieService) LoadCookies(ctx context.Context, email, username string) ([]*http.Cookie, error) {
	cookiesJSON, err := s.repo.GetCookies(ctx, email, username)
	if err != nil {
		return nil, err
	}
	var cookies []*http.Cookie
	err = json.Unmarshal(cookiesJSON, &cookies)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookies: %w", err)
	}
	return cookies, nil
}
