package sqlrepo

import (
	"context"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

// ProcessedMarkRepo provides database operations for processed marks
type ProcessedMarkRepo struct {
	db *sql.DB
}

// NewProcessedMarkRepo creates a new processed mark repository
func NewProcessedMarkRepo(db *sql.DB) *ProcessedMarkRepo {
	return &ProcessedMarkRepo{db: db}
}

// IsProcessed checks if a mark has been processed for a specific key and type
func (r *ProcessedMarkRepo) IsProcessed(ctx context.Context, key string, typ string) (bool, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var count int64
	err := db.WithContext(ctx).
		Model(&model.ProcessedMark{}).
		Where("key = ? AND type = ?", key, typ).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// MarkProcessed marks a mark as processed for a specific key and type
func (r *ProcessedMarkRepo) MarkProcessed(ctx context.Context, key string, typ string) error {
	db := sql.GetDBOrTx(ctx, r.db)
	processedMark := &model.ProcessedMark{
		Key:  key,
		Type: typ,
	}

	return db.WithContext(ctx).Create(processedMark).Error
}
