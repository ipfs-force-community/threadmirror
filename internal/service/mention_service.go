package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"gorm.io/gorm"
)

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
	NumTweets       int           `json:"numTweets"`
	Status          string        `json:"status"`
}

// MentionRepoInterface defines the interface for mention repo operations
type MentionRepoInterface interface {
	// Mention CRUD
	GetMentionByID(ctx context.Context, id string) (*model.Mention, error)
	GetMentionByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Mention, error)
	CreateMention(ctx context.Context, mention *model.Mention) error
	GetMentions(
		ctx context.Context,
		userID string,
		limit, offset int,
	) ([]model.Mention, int64, error)
	GetMentionsByUser(ctx context.Context, userID string, limit, offset int) ([]model.Mention, int64, error)
	UpdateMention(ctx context.Context, mention *model.Mention) error
}

// MentionService provides business logic for mention operations
type MentionService struct {
	mentionRepo MentionRepoInterface
	llm         llm.Model
	storage     ipfs.Storage
	threadRepo  ThreadRepoInterface
	db          *sql.DB
}

// NewMentionService creates a new mention service
func NewMentionService(
	mentionRepo MentionRepoInterface,
	llm llm.Model,
	storage ipfs.Storage,
	threadRepo ThreadRepoInterface,
	db *sql.DB,
) *MentionService {
	return &MentionService{
		mentionRepo: mentionRepo,
		llm:         llm,
		storage:     storage,
		threadRepo:  threadRepo,
		db:          db,
	}
}

// CreateMention creates a mention record and creates a pending thread
func (s *MentionService) CreateMention(
	ctx context.Context,
	userID, threadID string,
	mentionID *string,
	mentionCreateAt time.Time,
) (*MentionSummary, error) {
	var result *model.Mention

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Inject transaction into context for repo operations
		ctx := sql.WithTxToContext(ctx, tx)

		// Check if mention already exists for this user and thread
		mention, err := s.mentionRepo.GetMentionByUserIDAndThreadID(ctx, userID, threadID)
		if err != nil {
			if errors.Is(err, errutil.ErrNotFound) {
				mention = nil
			} else {
				return err
			}
		}
		if mention != nil {
			// Mention exists, Thread is already preloaded by repo
			result = mention
			return nil
		}

		// Create or get thread in pending status
		thread, err := s.threadRepo.GetThreadByID(ctx, threadID)
		if err != nil {
			if errors.Is(err, errutil.ErrNotFound) {
				// Create new thread in pending status
				thread = &model.Thread{
					ID:        threadID,
					Summary:   "Scraping in progress...",
					CID:       "",
					NumTweets: 0,
					Status:    model.ThreadStatusPending,
					// Author fields will be filled when scraping completes
					AuthorID:              "",
					AuthorName:            "",
					AuthorScreenName:      "",
					AuthorProfileImageURL: "",
				}
				err = s.threadRepo.CreateThread(ctx, thread)
				if err != nil {
					return fmt.Errorf("failed to create pending thread: %w", err)
				}
			} else {
				return fmt.Errorf("failed to check thread: %w", err)
			}
		}

		// Create mention record
		mention = &model.Mention{
			UserID:          userID,
			ThreadID:        threadID,
			MentionCreateAt: mentionCreateAt,
		}

		if mentionID != nil {
			mention.ID = *mentionID
		} else {
			mention.ID = threadID + "_" + userID // Composite ID for user mention of thread
		}

		if err := s.mentionRepo.CreateMention(ctx, mention); err != nil {
			return fmt.Errorf("failed to create mention: %w", err)
		}

		// Set the Thread association for the newly created mention
		mention.Thread = *thread
		result = mention
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.buildMentionSummary(result), nil
}

func (s *MentionService) GetMentionByID(ctx context.Context, id string) (*MentionSummary, error) {
	mention, err := s.mentionRepo.GetMentionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.buildMentionSummary(mention), nil
}

// GetMentions retrieves mentions based on feed type
func (s *MentionService) GetMentions(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]MentionSummary, int64, error) {
	mentions, total, err := s.mentionRepo.GetMentions(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tweets: %w", err)
	}

	// Build summaries using preloaded Thread associations
	mentionSummaries := make([]MentionSummary, 0, len(mentions))
	for _, mention := range mentions {
		mentionSummaries = append(mentionSummaries, *s.buildMentionSummary(&mention))
	}

	return mentionSummaries, total, nil
}

// buildMentionSummary builds a MentionSummary from a model.Mention with preloaded Thread
func (s *MentionService) buildMentionSummary(mention *model.Mention) *MentionSummary {
	var author *ThreadAuthor
	thread := &mention.Thread

	if thread.ID != "" && thread.AuthorID != "" {
		author = &ThreadAuthor{
			ID:              thread.AuthorID,
			Name:            thread.AuthorName,
			ScreenName:      thread.AuthorScreenName,
			ProfileImageURL: thread.AuthorProfileImageURL,
		}
	}

	contentPreview := ""
	NumTweets := 0
	cid := ""
	if thread.ID != "" {
		contentPreview = thread.Summary
		NumTweets = thread.NumTweets
		cid = thread.CID
	}

	return &MentionSummary{
		ID:              mention.ID,
		CID:             cid,
		ContentPreview:  contentPreview, // Use thread summary as content preview
		ThreadAuthor:    author,
		ThreadID:        mention.ThreadID,
		CreatedAt:       mention.CreatedAt,
		MentionCreateAt: mention.MentionCreateAt,
		NumTweets:       NumTweets,
		Status:          string(thread.Status),
	}
}
