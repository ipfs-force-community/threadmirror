package service

import (
	"context"
	"errors"
	"fmt"
)

// Processed Mention Service Errors
var (
	ErrMentionAlreadyProcessed = errors.New("mention already processed")
)

// ProcessedMarkRepoInterface defines the interface for processed mark repo operations
type ProcessedMarkRepoInterface interface {
	IsProcessed(ctx context.Context, key, typ string) (bool, error)
	MarkProcessed(ctx context.Context, key, typ string) error
}

// ProcessedMarkService provides business logic for tracking processed marks
type ProcessedMarkService struct {
	repo ProcessedMarkRepoInterface
}

// NewProcessedMarkService creates a new processed mark service
func NewProcessedMarkService(repo ProcessedMarkRepoInterface) *ProcessedMarkService {
	return &ProcessedMarkService{
		repo: repo,
	}
}

// IsProcessed checks if a mark has been processed for a specific key and type
func (s *ProcessedMarkService) IsProcessed(ctx context.Context, key, typ string) (bool, error) {
	processed, err := s.repo.IsProcessed(ctx, key, typ)
	if err != nil {
		return false, fmt.Errorf("failed to check if mark is processed: %w", err)
	}
	return processed, nil
}

// MarkProcessed marks a mark as processed for a specific key and type
func (s *ProcessedMarkService) MarkProcessed(ctx context.Context, key, typ string) error {
	// Check if already processed to avoid duplicates
	processed, err := s.repo.IsProcessed(ctx, key, typ)
	if err != nil {
		return fmt.Errorf("failed to check if mark is processed: %w", err)
	}

	if processed {
		return ErrMentionAlreadyProcessed
	}

	err = s.repo.MarkProcessed(ctx, key, typ)
	if err != nil {
		return fmt.Errorf("failed to mark as processed: %w", err)
	}

	return nil
}
