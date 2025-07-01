package sqlrepo

import (
	"context"
	"encoding/json"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"gorm.io/datatypes"
)

// BotCookieRepo provides database operations for bot cookies
type BotCookieRepo struct{}

// NewBotCookieRepo creates a new bot cookie repository
func NewBotCookieRepo() *BotCookieRepo {
	return &BotCookieRepo{}
}

// GetCookies retrieves cookies for the bot
func (r *BotCookieRepo) GetCookies(ctx context.Context, email, username string) (datatypes.JSON, error) {
	db := sql.MustDBFromContext(ctx)
	var botCookie model.BotCookie
	err := db.WithContext(ctx).
		Where("email = ? AND username = ?", email, username).
		First(&botCookie).Error

	if err != nil {
		return nil, err
	}

	return botCookie.CookiesData, nil
}

// SaveCookies saves or updates cookies for the bot
func (r *BotCookieRepo) SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error {
	db := sql.MustDBFromContext(ctx)
	// Convert cookies data to JSON
	jsonData, err := json.Marshal(cookiesData)
	if err != nil {
		return err
	}

	cookiesJSON := datatypes.JSON(jsonData)

	// Try to find existing record
	var existingCookie model.BotCookie
	err = db.WithContext(ctx).
		Where("email = ? AND username = ?", email, username).
		First(&existingCookie).Error

	if err != nil {
		// Record doesn't exist, create new one
		newCookie := &model.BotCookie{
			Email:       email,
			Username:    username,
			CookiesData: cookiesJSON,
		}
		return db.WithContext(ctx).Create(newCookie).Error
	}

	// Update existing record
	return db.WithContext(ctx).
		Model(&existingCookie).
		Update("cookies_data", cookiesJSON).Error
}
