package sqlrepo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BotCookieRepo provides database operations for bot cookies
type BotCookieRepo struct {
	db *sql.DB
}

// NewBotCookieRepo creates a new bot cookie repository
func NewBotCookieRepo(db *sql.DB) *BotCookieRepo {
	return &BotCookieRepo{db: db}
}

// GetCookies retrieves cookies for the bot
func (r *BotCookieRepo) GetCookies(ctx context.Context, email, username string) (datatypes.JSON, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var botCookie model.BotCookie
	err := db.WithContext(ctx).
		Where("email = ? AND username = ?", email, username).
		First(&botCookie).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errutil.ErrNotFound
		}
		return nil, err
	}

	return botCookie.CookiesData, nil
}

// SaveCookies saves or updates cookies for the bot
func (r *BotCookieRepo) SaveCookies(ctx context.Context, email, username string, cookiesData interface{}) error {
	db := sql.GetDBOrTx(ctx, r.db)
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

// GetLatestCookies retrieves the cookies data from the most recently updated record
func (r *BotCookieRepo) GetLatestBotCookie(ctx context.Context) (*model.BotCookie, error) {
	db := sql.GetDBOrTx(ctx, r.db)

	var botCookie model.BotCookie
	err := db.WithContext(ctx).
		Order("updated_at DESC").
		First(&botCookie).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errutil.ErrNotFound
		}
		return nil, err
	}

	return &botCookie, nil
}
