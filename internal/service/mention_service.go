package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs-force-community/threadmirror/internal/sqlc_generated"
	dbsql "github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/jackc/pgx/v5"
)

// Mention Service Errors - removed, using centralized errors.go

// MentionAuthor represents the author information in mentions
type ThreadAuthor struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ScreenName      string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

// MentionSummary represents a mention summary for list views
type MentionSummary struct {
	ID              string        `json:"id"`
	CID             string        `json:"cid"`
	ContentPreview  string        `json:"content_preview"`
	ThreadID        string        `json:"thread_id"`
	ThreadAuthor    *ThreadAuthor `json:"thread_author"`
	CreatedAt       time.Time     `json:"created_at"`
	MentionCreateAt time.Time     `json:"mention_create_at"`
	NumTweets       int           `json:"num_tweets"`
	Status          string        `json:"status"`
	RetryCount      int           `json:"retry_count"`
}

// MentionService provides business logic for mention operations
type MentionService struct {
	db      *dbsql.DB
	llm     llm.Model
	storage ipfs.Storage
}

// NewMentionService creates a new mention service
func NewMentionService(
	db *dbsql.DB,
	llm llm.Model,
	storage ipfs.Storage,
) *MentionService {
	return &MentionService{
		db:      db,
		llm:     llm,
		storage: storage,
	}
}

// CreateMention creates a mention record and creates a pending thread
func (s *MentionService) CreateMention(
	ctx context.Context,
	userID, threadID string,
	mentionID *string,
	mentionCreateAt time.Time,
) (*MentionSummary, error) {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return nil, fmt.Errorf("invalid thread ID: %w", err)
	}

	var mention sqlc_generated.Mention
	var thread sqlc_generated.Thread

	err = s.db.RunInTx(ctx, func(ctx context.Context) error {
		queries := s.db.QueriesFromContext(ctx)

		// Check if mention already exists for this user and thread
		_, err := queries.GetMentionByUserIDAndThreadID(ctx, sqlc_generated.GetMentionByUserIDAndThreadIDParams{
			UserID:   userID,
			ThreadID: threadUUID,
		})
		if err == nil {
			return ErrMentionAlreadyExists
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check existing mention: %w", err)
		}

		// Create or get thread in pending status
		thread, err = queries.GetThreadByID(ctx, sqlc_generated.GetThreadByIDParams{ThreadID: threadUUID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Create new thread in pending status
				thread, err = queries.CreateThread(ctx, sqlc_generated.CreateThreadParams{
					ID:        threadUUID,
					Summary:   "",
					Cid:       "",
					NumTweets: 0,
					Status:    "pending",
					// Author fields will be filled when scraping completes
				})
				if err != nil {
					return fmt.Errorf("failed to create pending thread: %w", err)
				}
			} else {
				return fmt.Errorf("failed to check thread: %w", err)
			}
		}

		// Create mention
		var mentionUUID uuid.UUID
		if mentionID != nil {
			mentionUUID, err = uuid.Parse(*mentionID)
			if err != nil {
				return fmt.Errorf("invalid mention ID: %w", err)
			}
		} else {
			mentionUUID = uuid.New()
		}

		mention, err = queries.CreateMention(ctx, sqlc_generated.CreateMentionParams{
			ID:              mentionUUID,
			UserID:          userID,
			ThreadID:        threadUUID,
			MentionCreateAt: mentionCreateAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create mention: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert to MentionSummary for response
	summary := &MentionSummary{
		ID:              mention.ID.String(),
		CID:             thread.Cid,
		ContentPreview:  thread.Summary,
		ThreadID:        thread.ID.String(),
		CreatedAt:       mention.CreatedAt,
		MentionCreateAt: mention.MentionCreateAt,
		NumTweets:       int(thread.NumTweets),
		Status:          thread.Status,
		RetryCount:      int(thread.RetryCount),
	}

	// Set thread author if available
	if thread.AuthorID != nil && *thread.AuthorID != "" {
		summary.ThreadAuthor = &ThreadAuthor{
			ID:              *thread.AuthorID,
			Name:            getStringValue(thread.AuthorName),
			ScreenName:      getStringValue(thread.AuthorScreenName),
			ProfileImageURL: getStringValue(thread.AuthorProfileImageUrl),
		}
	}

	return summary, nil
}

// GetMentions retrieves mentions with optional user filtering and pagination
func (s *MentionService) GetMentions(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]MentionSummary, int64, error) {
	// Get mentions with thread data
	mentionRows, err := s.db.QueriesFromContext(ctx).GetMentions(ctx, sqlc_generated.GetMentionsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("get mentions: %w", err)
	}

	// Get total count
	count, err := s.db.QueriesFromContext(ctx).CountMentions(ctx, sqlc_generated.CountMentionsParams{UserID: userID})
	if err != nil {
		return nil, 0, fmt.Errorf("count mentions: %w", err)
	}

	// Convert to MentionSummary
	summaries := make([]MentionSummary, len(mentionRows))
	for i, row := range mentionRows {
		summaries[i] = MentionSummary{
			ID:              row.ID.String(),
			CID:             row.Cid,
			ContentPreview:  row.Summary,
			ThreadID:        row.ThreadID.String(),
			CreatedAt:       row.CreatedAt,
			MentionCreateAt: row.MentionCreateAt,
			NumTweets:       int(row.NumTweets),
			Status:          row.Status,
			RetryCount:      int(row.RetryCount),
		}

		// Set thread author if available
		if row.AuthorID != nil && *row.AuthorID != "" {
			summaries[i].ThreadAuthor = &ThreadAuthor{
				ID:              *row.AuthorID,
				Name:            getStringValue(row.AuthorName),
				ScreenName:      getStringValue(row.AuthorScreenName),
				ProfileImageURL: getStringValue(row.AuthorProfileImageUrl),
			}
		}
	}

	return summaries, count, nil
}

// GetMentionsByUser retrieves mentions created by a specific user with pagination
func (s *MentionService) GetMentionsByUser(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]MentionSummary, int64, error) {
	// Get mentions with thread data
	mentionRows, err := s.db.QueriesFromContext(ctx).GetMentionsByUser(ctx, sqlc_generated.GetMentionsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("get mentions by user: %w", err)
	}

	// Get total count
	count, err := s.db.QueriesFromContext(ctx).CountMentionsByUser(ctx, sqlc_generated.CountMentionsByUserParams{UserID: userID})
	if err != nil {
		return nil, 0, fmt.Errorf("count mentions by user: %w", err)
	}

	// Convert to MentionSummary
	summaries := make([]MentionSummary, len(mentionRows))
	for i, row := range mentionRows {
		summaries[i] = MentionSummary{
			ID:              row.ID.String(),
			CID:             row.Cid,
			ContentPreview:  row.Summary,
			ThreadID:        row.ThreadID.String(),
			CreatedAt:       row.CreatedAt,
			MentionCreateAt: row.MentionCreateAt,
			NumTweets:       int(row.NumTweets),
			Status:          row.Status,
			RetryCount:      int(row.RetryCount),
		}

		// Set thread author if available
		if row.AuthorID != nil && *row.AuthorID != "" {
			summaries[i].ThreadAuthor = &ThreadAuthor{
				ID:              *row.AuthorID,
				Name:            getStringValue(row.AuthorName),
				ScreenName:      getStringValue(row.AuthorScreenName),
				ProfileImageURL: getStringValue(row.AuthorProfileImageUrl),
			}
		}
	}

	return summaries, count, nil
}

// GetMentionByID retrieves a mention by ID with thread data
func (s *MentionService) GetMentionByID(ctx context.Context, id string) (*MentionSummary, error) {
	mentionUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid mention ID: %w", err)
	}

	row, err := s.db.QueriesFromContext(ctx).GetMentionByID(ctx, sqlc_generated.GetMentionByIDParams{MentionID: mentionUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get mention by ID: %w", err)
	}

	summary := &MentionSummary{
		ID:              row.ID.String(),
		CID:             row.Cid,
		ContentPreview:  row.Summary,
		ThreadID:        row.ThreadID.String(),
		CreatedAt:       row.CreatedAt,
		MentionCreateAt: row.MentionCreateAt,
		NumTweets:       int(row.NumTweets),
		Status:          row.Status,
		RetryCount:      int(row.RetryCount),
	}

	// Set thread author if available
	if row.AuthorID != nil && *row.AuthorID != "" {
		summary.ThreadAuthor = &ThreadAuthor{
			ID:              *row.AuthorID,
			Name:            getStringValue(row.AuthorName),
			ScreenName:      getStringValue(row.AuthorScreenName),
			ProfileImageURL: getStringValue(row.AuthorProfileImageUrl),
		}
	}

	return summary, nil
}

// Helper function to safely get string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
