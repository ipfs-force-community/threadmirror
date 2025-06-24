package sqlrepo

import (
	"context"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/samber/lo"
)

// ProcessedMentionRepo provides database operations for processed mentions
type ProcessedMentionRepo struct {
	db *sql.DB
}

// NewProcessedMentionRepo creates a new processed mention repository
func NewProcessedMentionRepo(db *sql.DB) *ProcessedMentionRepo {
	return &ProcessedMentionRepo{db: db}
}

// IsProcessed checks if a mention has been processed for a specific user
func (r *ProcessedMentionRepo) IsProcessed(ctx context.Context, userID string, tweetID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ProcessedMention{}).
		Where("user_id = ? AND tweet_id = ?", userID, tweetID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// MarkProcessed marks a mention as processed for a specific user
func (r *ProcessedMentionRepo) MarkProcessed(ctx context.Context, userID string, tweetID string) error {
	processedMention := &model.ProcessedMention{
		UserID:  userID,
		TweetID: tweetID,
	}

	return r.db.WithContext(ctx).Create(processedMention).Error
}

// BatchMarkProcessed marks multiple mentions as processed for a specific user
func (r *ProcessedMentionRepo) BatchMarkProcessed(ctx context.Context, userID string, tweetIDs []string) error {
	if len(tweetIDs) == 0 {
		return nil
	}

	processedMentions := lo.Map(tweetIDs, func(id string, _ int) *model.ProcessedMention {
		return &model.ProcessedMention{
			UserID:  userID,
			TweetID: id,
		}
	})

	return r.db.WithContext(ctx).CreateInBatches(processedMentions, 100).Error
}
