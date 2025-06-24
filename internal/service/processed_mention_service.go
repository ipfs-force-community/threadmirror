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

// ProcessedMentionRepoInterface defines the interface for processed mention repo operations
type ProcessedMentionRepoInterface interface {
	IsProcessed(ctx context.Context, userID string, tweetID string) (bool, error)
	MarkProcessed(ctx context.Context, userID string, tweetID string) error
	BatchMarkProcessed(ctx context.Context, userID string, tweetIDs []string) error
}

// ProcessedMentionService provides business logic for tracking processed mentions
type ProcessedMentionService struct {
	repo ProcessedMentionRepoInterface
}

// NewProcessedMentionService creates a new processed mention service
func NewProcessedMentionService(repo ProcessedMentionRepoInterface) *ProcessedMentionService {
	return &ProcessedMentionService{
		repo: repo,
	}
}

// IsProcessed checks if a mention has been processed for a specific user
func (s *ProcessedMentionService) IsProcessed(ctx context.Context, userID string, tweetID string) (bool, error) {
	processed, err := s.repo.IsProcessed(ctx, userID, tweetID)
	if err != nil {
		return false, fmt.Errorf("failed to check if mention is processed: %w", err)
	}
	return processed, nil
}

// MarkProcessed marks a mention as processed for a specific user
func (s *ProcessedMentionService) MarkProcessed(ctx context.Context, userID string, tweetID string) error {
	// Check if already processed to avoid duplicates
	processed, err := s.repo.IsProcessed(ctx, userID, tweetID)
	if err != nil {
		return fmt.Errorf("failed to check if mention is processed: %w", err)
	}

	if processed {
		return ErrMentionAlreadyProcessed
	}

	err = s.repo.MarkProcessed(ctx, userID, tweetID)
	if err != nil {
		return fmt.Errorf("failed to mark mention as processed: %w", err)
	}

	return nil
}

// BatchMarkProcessed marks multiple mentions as processed for a specific user
func (s *ProcessedMentionService) BatchMarkProcessed(ctx context.Context, userID string, tweetIDs []string) error {
	if len(tweetIDs) == 0 {
		return nil
	}

	err := s.repo.BatchMarkProcessed(ctx, userID, tweetIDs)
	if err != nil {
		return fmt.Errorf("failed to batch mark mentions as processed: %w", err)
	}

	return nil
}
