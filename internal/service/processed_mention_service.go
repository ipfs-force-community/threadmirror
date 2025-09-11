package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/sqlc_generated"
	dbsql "github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	
	"github.com/jackc/pgx/v5"
)

type ProcessedMarkService struct {
	db *dbsql.DB
}

func NewProcessedMarkService(db *dbsql.DB) *ProcessedMarkService {
	return &ProcessedMarkService{db: db}
}

// IsProcessed checks if a key-type combination has been processed
func (s *ProcessedMarkService) IsProcessed(ctx context.Context, key, markType string) (bool, error) {
	_, err := s.db.QueriesFromContext(ctx).GetProcessedMark(ctx, sqlc_generated.GetProcessedMarkParams{
		Key:  key,
		Type: markType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get processed mark: %w", err)
	}
	return true, nil
}

// MarkAsProcessed marks a key-type combination as processed
func (s *ProcessedMarkService) MarkAsProcessed(ctx context.Context, key, markType string) error {
	_, err := s.db.QueriesFromContext(ctx).UpsertProcessedMark(ctx, sqlc_generated.UpsertProcessedMarkParams{
		Key:  key,
		Type: markType,
	})
	if err != nil {
		return fmt.Errorf("upsert processed mark: %w", err)
	}
	return nil
}

// DeleteProcessedMark removes a processed mark
func (s *ProcessedMarkService) DeleteProcessedMark(ctx context.Context, key, markType string) error {
	err := s.db.QueriesFromContext(ctx).DeleteProcessedMark(ctx, sqlc_generated.DeleteProcessedMarkParams{
		Key:  key,
		Type: markType,
	})
	if err != nil {
		return fmt.Errorf("delete processed mark: %w", err)
	}
	return nil
}

// CleanupOldMarks removes processed marks older than the specified duration
func (s *ProcessedMarkService) CleanupOldMarks(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	err := s.db.QueriesFromContext(ctx).DeleteOldProcessedMarks(ctx, sqlc_generated.DeleteOldProcessedMarksParams{CutoffTime: cutoffTime})
	if err != nil {
		return fmt.Errorf("cleanup old processed marks: %w", err)
	}
	return nil
}

// GetProcessedMark retrieves a specific processed mark
func (s *ProcessedMarkService) GetProcessedMark(ctx context.Context, key, markType string) (*sqlc_generated.ProcessedMark, error) {
	mark, err := s.db.QueriesFromContext(ctx).GetProcessedMark(ctx, sqlc_generated.GetProcessedMarkParams{
		Key:  key,
		Type: markType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get processed mark: %w", err)
	}
	return &mark, nil
}
