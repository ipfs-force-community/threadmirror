package sqlrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"gorm.io/gorm"
)

type ThreadRepo struct {
	db *sql.DB
}

func NewThreadRepo(db *sql.DB) *ThreadRepo {
	return &ThreadRepo{db: db}
}

func (r *ThreadRepo) GetThreadByID(ctx context.Context, id string) (*model.Thread, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var thread model.Thread
	err := db.Where("id = ?", id).First(&thread).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errutil.ErrNotFound
		}
		return nil, err
	}
	return &thread, nil
}

func (r *ThreadRepo) CreateThread(ctx context.Context, thread *model.Thread) error {
	db := sql.GetDBOrTx(ctx, r.db)
	return db.Create(thread).Error
}

// UpdateThread updates an existing thread with optimistic locking
func (r *ThreadRepo) UpdateThread(ctx context.Context, thread *model.Thread) error {
	db := sql.GetDBOrTx(ctx, r.db)

	// Store the current version to check for optimistic locking
	currentVersion := thread.Version

	// Increment version for optimistic locking
	thread.Version++

	// Update with version check
	result := db.Model(thread).
		Where("id = ? AND version = ?", thread.ID, currentVersion).
		Updates(thread)

	if result.Error != nil {
		// Restore original version on error
		thread.Version = currentVersion
		return result.Error
	}

	// If no rows affected, it's an optimistic lock conflict
	if result.RowsAffected == 0 {
		// Restore original version on conflict
		thread.Version = currentVersion
		return fmt.Errorf("optimistic lock conflict")
	}

	return nil
}

// UpdateThreadStatus updates thread status using optimistic locking
func (r *ThreadRepo) UpdateThreadStatus(ctx context.Context, threadID string, status model.ThreadStatus, version int) error {
	db := sql.GetDBOrTx(ctx, r.db)

	// First check if thread exists
	var exists bool
	err := db.Model(&model.Thread{}).Select("1").Where("id = ?", threadID).Scan(&exists).Error
	if err != nil {
		return err
	}
	if !exists {
		return errutil.ErrNotFound
	}

	// Atomic update with version check for optimistic locking
	result := db.Model(&model.Thread{}).
		Where("id = ? AND version = ?", threadID, version).
		Updates(map[string]interface{}{
			"status":  status,
			"version": version + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	// If no rows affected, it must be version conflict (we already confirmed thread exists)
	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock conflict")
	}

	return nil
}

func (r *ThreadRepo) GetTweetsByIDs(ctx context.Context, ids []string) (map[string]*model.Thread, error) {
	if len(ids) == 0 {
		return map[string]*model.Thread{}, nil
	}
	db := sql.GetDBOrTx(ctx, r.db)
	var tweets []model.Thread
	err := db.Where("id IN ?", ids).Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]*model.Thread, len(tweets))
	for i := range tweets {
		result[tweets[i].ID] = &tweets[i]
	}
	return result, nil
}

// GetStuckScrapingThreadsForRetry finds threads stuck in 'scraping' status, increments their retry count, and returns them
func (r *ThreadRepo) GetStuckScrapingThreadsForRetry(ctx context.Context, stuckDuration time.Duration, maxRetries int) ([]*model.Thread, error) {
	db := sql.GetDBOrTx(ctx, r.db)

	cutoffTime := time.Now().Add(-stuckDuration)

	// Execute in transaction to ensure atomicity
	var threads []*model.Thread
	err := db.Transaction(func(tx *gorm.DB) error {
		// First, find threads that match criteria
		var threadIDs []string
		err := tx.Model(&model.Thread{}).
			Select("id").
			Where("status = ? AND updated_at < ? AND retry_count < ?",
				model.ThreadStatusScraping, cutoffTime, maxRetries).
			Find(&threadIDs).Error
		if err != nil {
			return err
		}

		if len(threadIDs) == 0 {
			return nil // No threads found
		}

		// Increment retry_count for found threads
		err = tx.Model(&model.Thread{}).
			Where("id IN ?", threadIDs).
			Update("retry_count", gorm.Expr("retry_count + 1")).Error
		if err != nil {
			return err
		}

		// Return the updated threads
		return tx.Where("id IN ?", threadIDs).Find(&threads).Error
	})

	return threads, err
}

// GetOldPendingThreadsForRetry finds threads that have been pending too long, increments their retry count, and returns them
func (r *ThreadRepo) GetOldPendingThreadsForRetry(ctx context.Context, pendingDuration time.Duration, maxRetries int) ([]*model.Thread, error) {
	db := sql.GetDBOrTx(ctx, r.db)

	cutoffTime := time.Now().Add(-pendingDuration)

	// Execute in transaction to ensure atomicity
	var threads []*model.Thread
	err := db.Transaction(func(tx *gorm.DB) error {
		// First, find threads that match criteria
		var threadIDs []string
		err := tx.Model(&model.Thread{}).
			Select("id").
			Where("status = ? AND created_at < ? AND retry_count < ?",
				model.ThreadStatusPending, cutoffTime, maxRetries).
			Find(&threadIDs).Error
		if err != nil {
			return err
		}

		if len(threadIDs) == 0 {
			return nil // No threads found
		}

		// Increment retry_count for found threads
		err = tx.Model(&model.Thread{}).
			Where("id IN ?", threadIDs).
			Update("retry_count", gorm.Expr("retry_count + 1")).Error
		if err != nil {
			return err
		}

		// Return the updated threads
		return tx.Where("id IN ?", threadIDs).Find(&threads).Error
	})

	return threads, err
}

// GetFailedThreadsForRetry finds failed threads that can be retried, increments their retry count, and returns them
func (r *ThreadRepo) GetFailedThreadsForRetry(ctx context.Context, retryDelay time.Duration, maxRetries int) ([]*model.Thread, error) {
	db := sql.GetDBOrTx(ctx, r.db)

	cutoffTime := time.Now().Add(-retryDelay)

	// Execute in transaction to ensure atomicity
	var threads []*model.Thread
	err := db.Transaction(func(tx *gorm.DB) error {
		// First, find threads that match criteria
		var threadIDs []string
		err := tx.Model(&model.Thread{}).
			Select("id").
			Where("status = ? AND updated_at < ? AND retry_count < ?",
				model.ThreadStatusFailed, cutoffTime, maxRetries).
			Find(&threadIDs).Error
		if err != nil {
			return err
		}

		if len(threadIDs) == 0 {
			return nil // No threads found
		}

		// Increment retry_count for found threads
		err = tx.Model(&model.Thread{}).
			Where("id IN ?", threadIDs).
			Update("retry_count", gorm.Expr("retry_count + 1")).Error
		if err != nil {
			return err
		}

		// Return the updated threads
		return tx.Where("id IN ?", threadIDs).Find(&threads).Error
	})

	return threads, err
}
