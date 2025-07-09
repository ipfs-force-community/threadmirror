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
type MentionRepo struct {
	db *sql.DB
}

// NewMentionRepo creates a new mention repo
func NewMentionRepo(db *sql.DB) *MentionRepo {
	return &MentionRepo{db: db}
}

// Mention CRUD operations

// GetMentionByID retrieves a mention by ID with Thread preloaded
func (r *MentionRepo) GetMentionByID(ctx context.Context, id string) (*model.Mention, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var mention model.Mention
	err := db.Preload("Thread").Where("id = ?", id).First(&mention).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errutil.ErrNotFound
		}
		return nil, err
	}
	return &mention, nil
}

func (r *MentionRepo) GetMentionByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Mention, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var mention model.Mention
	err := db.Preload("Thread").Where("user_id = ? AND thread_id = ?", userID, threadID).First(&mention).Error
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
	db := sql.GetDBOrTx(ctx, r.db)
	return db.Create(mention).Error
}

// UpdateMention updates an existing mention
func (r *MentionRepo) UpdateMention(ctx context.Context, mention *model.Mention) error {
	db := sql.GetDBOrTx(ctx, r.db)
	return db.Save(mention).Error
}

// GetMentions retrieves mentions based on feed type with optional filtering
func (r *MentionRepo) GetMentions(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]model.Mention, int64, error) {
	db := sql.GetDBOrTx(ctx, r.db)
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

	// Get paginated results with Thread preloaded
	err := query.Preload("Thread").Order("mentions.created_at DESC").Limit(limit).Offset(offset).Find(&mentions).Error
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
	db := sql.GetDBOrTx(ctx, r.db)
	var mentions []model.Mention
	var total int64

	query := db.Model(&model.Mention{}).
		Where("user_id = ?", userID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with Thread preloaded
	err := query.Preload("Thread").Order("created_at DESC").Limit(limit).Offset(offset).Find(&mentions).Error
	if err != nil {
		return nil, 0, err
	}

	return mentions, total, nil
}
