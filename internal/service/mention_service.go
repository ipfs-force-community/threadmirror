package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/internal/repo/sqlrepo"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs/go-cid"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/gorm"
)

// Mention Service Errors
var (
	ErrMentionNotFound = errors.New("mention not found")
)

// MentionAuthor represents the author information in mentions
type MentionAuthor struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ScreenName      string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

// MentionDetail represents a complete mention with all details
type MentionDetail struct {
	ID        string            `json:"id"`
	CID       string            `json:"cid"`
	Tweets    []*xscraper.Tweet `json:"tweets,omitempty"`
	Author    *MentionAuthor    `json:"author"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// MentionSummary represents a mention summary for list views
type MentionSummary struct {
	ID             string         `json:"id"`
	CID            string         `json:"cid"`
	ContentPreview string         `json:"content_preview"`
	Author         *MentionAuthor `json:"author"`
	CreatedAt      time.Time      `json:"created_at"`
	NumTweets      int            `json:"numTweets"`
}

// CreateMentionRequest represents a request to create a new mention
type CreateMentionRequest struct {
	Tweets []*xscraper.Tweet
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
}

// MentionService provides business logic for mention operations
type MentionService struct {
	mentionRepo MentionRepoInterface
	llm         llm.Model
	storage     ipfs.Storage
	threadRepo  *sqlrepo.ThreadRepo
	db          *sql.DB
}

// NewMentionService creates a new mention service
func NewMentionService(
	mentionRepo MentionRepoInterface,
	llm llm.Model,
	storage ipfs.Storage,
	threadRepo *sqlrepo.ThreadRepo,
) *MentionService {
	return &MentionService{
		mentionRepo: mentionRepo,
		llm:         llm,
		storage:     storage,
		threadRepo:  threadRepo,
	}
}

// CreateMention creates a new mention
func (s *MentionService) CreateMention(
	ctx context.Context,
	req *CreateMentionRequest,
) (*MentionDetail, error) {
	var result *MentionDetail
	db := sql.MustDBFromContext(ctx)
	return result, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		mentionRepo := s.mentionRepo
		threadRepo := s.threadRepo

		if len(req.Tweets) < 2 {
			return fmt.Errorf("no tweets provided")
		}

		threadID := req.Tweets[len(req.Tweets)-2].RestID
		mentionTweet := req.Tweets[len(req.Tweets)-1]

		// 去重逻辑：如已存在则直接返回
		mention, err := mentionRepo.GetMentionByUserIDAndThreadID(ctx, mentionTweet.Author.RestID, threadID)
		if err != nil {
			return err
		}
		if mention != nil {
			result, err = s.buildMentionDetail(ctx, mention)
			if err != nil {
				return err
			}
			return nil
		}

		_, err = threadRepo.GetThreadByID(ctx, threadID)
		if err != nil {
			return fmt.Errorf("failed to check thread existence: %w", err)
		}

		jsonThread, err := json.Marshal(req.Tweets[:len(req.Tweets)-1])
		if err != nil {
			return fmt.Errorf("failed to marshal tweets: %w", err)
		}

		summary, err := s.generateTweetsSummary(ctx, string(jsonThread))
		if err != nil {
			return fmt.Errorf("failed to generate AI summary: %w", err)
		}

		cid, err := s.storage.Add(ctx, bytes.NewReader(jsonThread))
		if err != nil {
			return fmt.Errorf("failed to add tweets to IPFS: %w", err)
		}

		var authorID, authorName, authorScreenName, authorProfileImageURL string
		if req.Tweets[len(req.Tweets)-2].Author != nil {
			author := req.Tweets[len(req.Tweets)-2].Author
			authorID = author.RestID
			authorName = author.Name
			authorScreenName = author.ScreenName
			authorProfileImageURL = author.ProfileImageURL
		}
		err = threadRepo.CreateThread(ctx, &model.Thread{
			ID:                    threadID,
			Summary:               summary,
			CID:                   cid.String(),
			AuthorID:              authorID,
			AuthorName:            authorName,
			AuthorScreenName:      authorScreenName,
			AuthorProfileImageURL: authorProfileImageURL,
			NumTweets:             len(req.Tweets) - 1,
		})
		if err != nil {
			return fmt.Errorf("failed to create thread: %w", err)
		}

		mention = &model.Mention{
			ID:       mentionTweet.RestID,
			UserID:   mentionTweet.Author.RestID,
			ThreadID: threadID,
		}

		if err := mentionRepo.CreateMention(ctx, mention); err != nil {
			return fmt.Errorf("failed to create mention: %w", err)
		}

		// 事务内查详情
		md, err := s.GetMentionByID(ctx, mention.ID)
		if err != nil {
			return err
		}
		result = md
		return nil
	})
}

// GetMentionByID retrieves a mention by ID
func (s *MentionService) GetMentionByID(ctx context.Context, mentionID string) (*MentionDetail, error) {
	mention, err := s.mentionRepo.GetMentionByID(ctx, mentionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mention: %w", err)
	}
	if mention == nil {
		return nil, ErrMentionNotFound
	}
	return s.buildMentionDetail(ctx, mention)
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

	// 批量查 thread
	threadIDs := make([]string, 0, len(mentions))
	for _, mention := range mentions {
		if mention.ThreadID != "" {
			threadIDs = append(threadIDs, mention.ThreadID)
		}
	}
	tweetsMap := map[string]*model.Thread{}
	if len(threadIDs) > 0 && s.threadRepo != nil {
		tweetsMap, err = s.threadRepo.GetTweetsByIDs(ctx, threadIDs)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get threads: %w", err)
		}
	}

	mentionSummaries := make([]MentionSummary, 0, len(mentions))
	for _, mention := range mentions {
		thread := tweetsMap[mention.ThreadID]
		mentionSummaries = append(mentionSummaries, *s.buildMentionSummary(&mention, thread))
	}

	return mentionSummaries, total, nil
}

// buildMentionDetail builds a MentionDetail from a model.Mention
func (s *MentionService) buildMentionDetail(
	ctx context.Context,
	mention *model.Mention,
) (*MentionDetail, error) {
	threadRepo := sqlrepo.NewThreadRepo()
	thread, err := threadRepo.GetThreadByID(ctx, mention.ThreadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	if thread == nil {
		return nil, fmt.Errorf("thread not found")
	}

	// Build author information
	var author *MentionAuthor
	if thread.AuthorID != "" {
		author = &MentionAuthor{
			ID:              thread.AuthorID,
			Name:            thread.AuthorName,
			ScreenName:      thread.AuthorScreenName,
			ProfileImageURL: thread.AuthorProfileImageURL,
		}
	}

	// Load tweets from IPFS
	tweets, _ := s.loadTweetsFromIPFS(ctx, thread.CID)

	return &MentionDetail{
		ID:        mention.ID,
		CID:       thread.CID,
		Tweets:    tweets,
		Author:    author,
		CreatedAt: mention.CreatedAt,
		UpdatedAt: mention.UpdatedAt,
	}, nil
}

// buildMentionSummary builds a MentionSummary from a model.Mention
func (s *MentionService) buildMentionSummary(
	mention *model.Mention,
	thread *model.Thread,
) *MentionSummary {
	// Build author information
	var author *MentionAuthor
	if thread != nil && thread.AuthorID != "" {
		author = &MentionAuthor{
			ID:              thread.AuthorID,
			Name:            thread.AuthorName,
			ScreenName:      thread.AuthorScreenName,
			ProfileImageURL: thread.AuthorProfileImageURL,
		}
	}

	contentPreview := ""
	if thread != nil {
		contentPreview = thread.Summary
	}

	NumTweets := 0
	if thread != nil {
		NumTweets = thread.NumTweets
	}

	return &MentionSummary{
		ID: mention.ID,
		CID: func() string {
			if thread != nil {
				return thread.CID
			}
			return ""
		}(),
		ContentPreview: contentPreview, // Use thread summary as content preview
		Author:         author,
		CreatedAt:      mention.CreatedAt,
		NumTweets:      NumTweets,
	}
}

// generateTweetsSummary generates AI summary for tweets
func (s *MentionService) generateTweetsSummary(ctx context.Context, jsonTweets string) (string, error) {

	// Create prompt for AI summarization
	prompt := fmt.Sprintf(`Please analyze the following JSON data containing Twitter/X posts and provide a concise summary (maximum 200 characters) in Chinese. 

The JSON contains an array of tweet objects, each with fields like "text", "author", "created_at", etc. Focus on the main content and key themes from the "text" fields.

JSON Data:
%s

Please provide a Chinese summary:`, jsonTweets)

	// Generate summary using LLM
	summary, err := llms.GenerateFromSinglePrompt(ctx, s.llm, prompt,
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}
	summary = strings.TrimSpace(strings.ToValidUTF8(summary, ""))

	// Ensure summary is not too long
	if len([]rune(summary)) > 200 {
		runes := []rune(summary)
		summary = string(runes[:200]) + "..."
	}
	// Filter out invalid UTF-8 characters
	return strings.TrimSpace(summary), nil
}

// loadTweetsFromIPFS loads tweets from IPFS using the CID
func (s *MentionService) loadTweetsFromIPFS(ctx context.Context, cidStr string) ([]*xscraper.Tweet, error) {
	if cidStr == "" {
		return nil, nil
	}

	// Parse CID from string
	c, err := cid.Parse(cidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CID: %w", err)
	}

	// Get content from IPFS
	reader, err := s.storage.Get(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to get content from IPFS: %w", err)
	}
	defer reader.Close() // nolint:errcheck

	// Read all content
	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Unmarshal tweets
	var tweets []*xscraper.Tweet
	err = json.Unmarshal(buffer.Bytes(), &tweets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tweets: %w", err)
	}

	return tweets, nil
}

// CreateThreadMention creates a new mention
func (s *MentionService) CreateThreadMention(
	ctx context.Context,
	userID string,
	req *CreateMentionRequest,
) (*MentionDetail, error) {
	var result *MentionDetail
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx := sql.WithDBToContext(ctx, &sql.DB{DB: tx})
		mentionRepo := s.mentionRepo
		threadRepo := s.threadRepo

		if len(req.Tweets) < 2 {
			return fmt.Errorf("no tweets provided")
		}

		threadID := req.Tweets[len(req.Tweets)-2].RestID
		var err error
		_, err = threadRepo.GetThreadByID(ctx, threadID)
		if err != nil {
			return fmt.Errorf("failed to check thread existence: %w", err)
		}

		jsonThread, err := json.Marshal(req.Tweets[:len(req.Tweets)-1])
		if err != nil {
			return fmt.Errorf("failed to marshal tweets: %w", err)
		}

		summary, err := s.generateTweetsSummary(ctx, string(jsonThread))
		if err != nil {
			return fmt.Errorf("failed to generate AI summary: %w", err)
		}

		cid, err := s.storage.Add(ctx, bytes.NewReader(jsonThread))
		if err != nil {
			return fmt.Errorf("failed to add tweets to IPFS: %w", err)
		}

		var authorID, authorName, authorScreenName, authorProfileImageURL string
		if req.Tweets[len(req.Tweets)-2].Author != nil {
			author := req.Tweets[len(req.Tweets)-2].Author
			authorID = author.RestID
			authorName = author.Name
			authorScreenName = author.ScreenName
			authorProfileImageURL = author.ProfileImageURL
		}
		err = threadRepo.CreateThread(ctx, &model.Thread{
			ID:                    threadID,
			Summary:               summary,
			CID:                   cid.String(),
			AuthorID:              authorID,
			AuthorName:            authorName,
			AuthorScreenName:      authorScreenName,
			AuthorProfileImageURL: authorProfileImageURL,
			NumTweets:             len(req.Tweets) - 1,
		})
		if err != nil {
			return fmt.Errorf("failed to create thread: %w", err)
		}

		mention := &model.Mention{
			ID:       req.Tweets[len(req.Tweets)-1].RestID,
			UserID:   userID,
			ThreadID: threadID,
		}

		if err := mentionRepo.CreateMention(ctx, mention); err != nil {
			return fmt.Errorf("failed to create mention: %w", err)
		}

		// 事务内查详情，cctx 保证用事务 repo
		md, err := s.GetMentionByID(ctx, mention.ID)
		if err != nil {
			return err
		}
		result = md
		return nil
	})
	return result, err
}
