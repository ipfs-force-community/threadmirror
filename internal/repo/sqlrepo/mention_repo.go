package sqlrepo

import (
	"context"
	"errors"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"gorm.io/gorm"
)

// MentionRepo implements MentionRepoInterface
type MentionRepo struct{}

// NewMentionRepo creates a new mention repo
func NewMentionRepo() *MentionRepo {
	return &MentionRepo{}
}

// Mention CRUD operations

// GetMentionByID retrieves a mention by ID
func (r *MentionRepo) GetMentionByID(ctx context.Context, id string) (*model.Mention, error) {
	db := sql.MustDBFromContext(ctx)
	var mention model.Mention
	err := db.Where("id = ?", id).First(&mention).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errutil.ErrNotFound
		}
		return nil, err
	}
	return &mention, nil
}

func (r *MentionRepo) GetMentionByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Mention, error) {
	db := sql.MustDBFromContext(ctx)
	var mention model.Mention
	err := db.Where("user_id = ? AND thread_id = ?", userID, threadID).First(&mention).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errutil.ErrNotFound
		}
		return nil, err
	}
	return &mention, nil
}

// CreateMention creates a new mention
func (r *MentionRepo) CreateMention(ctx context.Context, mention *model.Mention) error {
	db := sql.MustDBFromContext(ctx)
	return db.Create(mention).Error
}

// GetMentions retrieves mentions based on feed type with optional filtering
func (r *MentionRepo) GetMentions(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]model.Mention, int64, error) {
	db := sql.MustDBFromContext(ctx)
	var mentions []model.Mention
	var total int64

	query := db.Model(&model.Mention{})

	// Filter by user if provided
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("mentions.created_at DESC").Limit(limit).Offset(offset).Find(&mentions).Error
	if err != nil {
		return nil, 0, err
	}

	return mentions, total, nil
}

// GetMentionsByUser retrieves mentions created by a specific user with pagination
func (r *MentionRepo) GetMentionsByUser(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]model.Mention, int64, error) {
	db := sql.MustDBFromContext(ctx)
	var mentions []model.Mention
	var total int64

	query := db.Model(&model.Mention{}).
		Where("user_id = ?", userID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&mentions).Error
	if err != nil {
		return nil, 0, err
	}

	return mentions, total, nil
}
