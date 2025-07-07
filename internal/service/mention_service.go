package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/llm"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/tmc/langchaingo/llms"
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
	threadRepo  ThreadRepoInterface
}

// NewMentionService creates a new mention service
func NewMentionService(
	mentionRepo MentionRepoInterface,
	llm llm.Model,
	storage ipfs.Storage,
	threadRepo ThreadRepoInterface,
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
) (*MentionSummary, error) {
	var result *model.Mention
	db := sql.MustDBFromContext(ctx)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
			result = mention
			return nil
		}

		thread, err := threadRepo.GetThreadByID(ctx, threadID)
		if err != nil {
			return fmt.Errorf("failed to check thread existence: %w", err)
		}
		if thread == nil {
			threadTweets := req.Tweets[:len(req.Tweets)-1]
			summary, err := s.generateTweetsSummary(ctx, threadTweets)
			if err != nil {
				return fmt.Errorf("failed to generate AI summary: %w", err)
			}

			jsonThread, err := json.Marshal(threadTweets)
			if err != nil {
				return fmt.Errorf("failed to marshal tweets: %w", err)
			}

			cid, err := s.storage.Add(ctx, bytes.NewReader(jsonThread))
			if err != nil {
				return fmt.Errorf("failed to add tweets to IPFS: %w", err)
			}

			err = threadRepo.CreateThread(ctx, &model.Thread{
				ID:        threadID,
				Summary:   summary,
				CID:       cid.String(),
				NumTweets: len(req.Tweets) - 1,
			})
			if err != nil {
				return fmt.Errorf("failed to create thread: %w", err)
			}
		}

		var authorID, authorName, authorScreenName, authorProfileImageURL string
		if threadAuthor := req.Tweets[len(req.Tweets)-2].Author; threadAuthor != nil {
			authorID = threadAuthor.RestID
			authorName = threadAuthor.Name
			authorScreenName = threadAuthor.ScreenName
			authorProfileImageURL = threadAuthor.ProfileImageURL
		}
		mention = &model.Mention{
			ID:                          mentionTweet.RestID,
			UserID:                      mentionTweet.Author.RestID,
			MentionCreateAt:             mentionTweet.CreatedAt,
			ThreadAuthorID:              authorID,
			ThreadAuthorName:            authorName,
			ThreadAuthorScreenName:      authorScreenName,
			ThreadAuthorProfileImageURL: authorProfileImageURL,
			ThreadID:                    threadID,
		}

		if err := mentionRepo.CreateMention(ctx, mention); err != nil {
			return fmt.Errorf("failed to create mention: %w", err)
		}

		// 事务内查详情
		md, err := mentionRepo.GetMentionByID(ctx, mention.ID)
		if err != nil {
			return err
		}
		result = md
		return nil
	})

	if err != nil {
		return nil, err
	}
	return s.buildMentionSummary(result, nil), nil
}

func (s *MentionService) GetMentionByID(ctx context.Context, id string) (*MentionSummary, error) {
	mention, err := s.mentionRepo.GetMentionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if mention == nil {
		return nil, nil
	}
	return s.buildMentionSummary(mention, nil), nil
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
	if len(threadIDs) > 0 {
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

// buildMentionSummary builds a MentionSummary from a model.Mention
func (s *MentionService) buildMentionSummary(
	mention *model.Mention,
	thread *model.Thread,
) *MentionSummary {
	author := &ThreadAuthor{
		ID:              mention.ThreadAuthorID,
		Name:            mention.ThreadAuthorName,
		ScreenName:      mention.ThreadAuthorScreenName,
		ProfileImageURL: mention.ThreadAuthorProfileImageURL,
	}

	contentPreview := ""
	NumTweets := 0
	cid := ""
	if thread != nil {
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
	}
}

// generateTweetsSummary generates AI summary for tweets
func (s *MentionService) generateTweetsSummary(ctx context.Context, tweets []*xscraper.Tweet) (string, error) {
	type ToSummarize struct {
		Text   string `json:"text"`
		Author struct {
			Name       string `json:"name"`
			ScreenName string `json:"screen_name"`
		} `json:"author"`
		IsRetweet bool `json:"is_retweet"`
		IsReply   bool `json:"is_reply"`
	}
	toSummarize := make([]ToSummarize, 0, len(tweets))
	for _, tweet := range tweets {
		toSummarize = append(toSummarize, ToSummarize{
			Text: tweet.Text,
			Author: struct {
				Name       string `json:"name"`
				ScreenName string `json:"screen_name"`
			}{
				Name:       tweet.Author.Name,
				ScreenName: tweet.Author.ScreenName,
			},
			IsRetweet: tweet.IsRetweet,
			IsReply:   tweet.IsReply,
		})
	}

	jsonTweets, err := json.Marshal(toSummarize)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tweets: %w", err)
	}

	// Create prompt for AI summarization
	prompt := fmt.Sprintf(`Please analyze the following JSON data containing Twitter/X posts and provide a concise summary (maximum 200 characters) in Chinese. 

The JSON contains an array of tweet objects, each with fields like "text", "author", etc. Focus on the main content and key themes from the "text" fields.

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
