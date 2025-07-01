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

// Post Service Errors
var (
	ErrPostNotFound = errors.New("post not found")
)

// PostAuthor represents the author information in posts
type PostAuthor struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ScreenName      string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

// PostDetail represents a complete post with all details
type PostDetail struct {
	ID        string            `json:"id"`
	CID       string            `json:"cid"`
	Tweets    []*xscraper.Tweet `json:"tweets,omitempty"`
	Author    *PostAuthor       `json:"author"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// PostSummary represents a post summary for list views
type PostSummary struct {
	ID             string      `json:"id"`
	CID            string      `json:"cid"`
	ContentPreview string      `json:"content_preview"`
	Author         *PostAuthor `json:"author"`
	CreatedAt      time.Time   `json:"created_at"`
	NumTweets      int         `json:"numTweets"`
}

// CreatePostRequest represents a request to create a new post
type CreatePostRequest struct {
	Tweets []*xscraper.Tweet
}

// PostRepoInterface defines the interface for post repo operations
type PostRepoInterface interface {
	// Post CRUD
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	GetPostByUserIDAndThreadID(ctx context.Context, userID, threadID string) (*model.Post, error)
	CreatePost(ctx context.Context, post *model.Post) error
	GetPosts(
		ctx context.Context,
		userID string,
		limit, offset int,
	) ([]model.Post, int64, error)
	GetPostsByUser(ctx context.Context, userID string, limit, offset int) ([]model.Post, int64, error)
}

// PostService provides business logic for post operations
type PostService struct {
	postRepo   PostRepoInterface
	llm        llm.Model
	storage    ipfs.Storage
	threadRepo *sqlrepo.ThreadRepo
	db         *sql.DB
}

// NewPostService creates a new post service
func NewPostService(
	postRepo PostRepoInterface,
	llm llm.Model,
	storage ipfs.Storage,
	threadRepo *sqlrepo.ThreadRepo,
) *PostService {
	return &PostService{
		postRepo:   postRepo,
		llm:        llm,
		storage:    storage,
		threadRepo: threadRepo,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(
	ctx context.Context,
	req *CreatePostRequest,
) (*PostDetail, error) {
	var result *PostDetail
	db := sql.MustDBFromContext(ctx)
	return result, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		postRepo := s.postRepo
		threadRepo := s.threadRepo

		if len(req.Tweets) < 2 {
			return fmt.Errorf("no tweets provided")
		}

		threadID := req.Tweets[len(req.Tweets)-2].RestID
		mentionTweet := req.Tweets[len(req.Tweets)-1]

		// 去重逻辑：如已存在则直接返回
		post, err := postRepo.GetPostByUserIDAndThreadID(ctx, mentionTweet.Author.RestID, threadID)
		if err != nil {
			return err
		}
		if post != nil {
			result, err = s.buildPostDetail(ctx, post)
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

		post = &model.Post{
			ID:       mentionTweet.RestID,
			UserID:   mentionTweet.Author.RestID,
			ThreadID: threadID,
		}

		if err := postRepo.CreatePost(ctx, post); err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}

		// 事务内查详情
		pd, err := s.GetPostByID(ctx, post.ID)
		if err != nil {
			return err
		}
		result = pd
		return nil
	})
}

// GetPostByID retrieves a post by ID
func (s *PostService) GetPostByID(ctx context.Context, postID string) (*PostDetail, error) {
	postRepo := sqlrepo.NewPostRepo()
	post, err := postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	return s.buildPostDetail(ctx, post)
}

// GetPosts retrieves posts based on feed type
func (s *PostService) GetPosts(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]PostSummary, int64, error) {
	posts, total, err := s.postRepo.GetPosts(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tweets: %w", err)
	}

	// 批量查 thread
	threadIDs := make([]string, 0, len(posts))
	for _, post := range posts {
		if post.ThreadID != "" {
			threadIDs = append(threadIDs, post.ThreadID)
		}
	}
	tweetsMap := map[string]*model.Thread{}
	if len(threadIDs) > 0 && s.threadRepo != nil {
		tweetsMap, err = s.threadRepo.GetTweetsByIDs(ctx, threadIDs)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get threads: %w", err)
		}
	}

	postSummaries := make([]PostSummary, 0, len(posts))
	for _, post := range posts {
		thread := tweetsMap[post.ThreadID]
		postSummaries = append(postSummaries, *s.buildPostSummary(&post, thread))
	}

	return postSummaries, total, nil
}

// buildPostDetail builds a PostDetail from a model.Post
func (s *PostService) buildPostDetail(
	ctx context.Context,
	post *model.Post,
) (*PostDetail, error) {
	threadRepo := sqlrepo.NewThreadRepo()
	thread, err := threadRepo.GetThreadByID(ctx, post.ThreadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	if thread == nil {
		return nil, fmt.Errorf("thread not found")
	}

	// Build author information
	var author *PostAuthor
	if thread.AuthorID != "" {
		author = &PostAuthor{
			ID:              thread.AuthorID,
			Name:            thread.AuthorName,
			ScreenName:      thread.AuthorScreenName,
			ProfileImageURL: thread.AuthorProfileImageURL,
		}
	}

	// Load tweets from IPFS
	tweets, _ := s.loadTweetsFromIPFS(ctx, thread.CID)

	return &PostDetail{
		ID:        post.ID,
		CID:       thread.CID,
		Tweets:    tweets,
		Author:    author,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

// buildPostSummary builds a PostSummaryDetail from a model.Post
func (s *PostService) buildPostSummary(
	post *model.Post,
	thread *model.Thread,
) *PostSummary {
	// Build author information
	var author *PostAuthor
	if thread != nil && thread.AuthorID != "" {
		author = &PostAuthor{
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

	return &PostSummary{
		ID: post.ID,
		CID: func() string {
			if thread != nil {
				return thread.CID
			}
			return ""
		}(),
		ContentPreview: contentPreview, // Use thread summary as content preview
		Author:         author,
		CreatedAt:      post.CreatedAt,
		NumTweets:      NumTweets,
	}
}

// generateTweetsSummary generates AI summary for tweets
func (s *PostService) generateTweetsSummary(ctx context.Context, jsonTweets string) (string, error) {

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
func (s *PostService) loadTweetsFromIPFS(ctx context.Context, cidStr string) ([]*xscraper.Tweet, error) {
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

// CreateThreadPost creates a new post
func (s *PostService) CreateThreadPost(
	ctx context.Context,
	userID string,
	req *CreatePostRequest,
) (*PostDetail, error) {
	var result *PostDetail
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx := sql.WithDBToContext(ctx, &sql.DB{DB: tx})
		postRepo := s.postRepo
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

		post := &model.Post{
			ID:       req.Tweets[len(req.Tweets)-1].RestID,
			UserID:   userID,
			ThreadID: threadID,
		}

		if err := postRepo.CreatePost(ctx, post); err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}

		// 事务内查详情，cctx 保证用事务 repo
		pd, err := s.GetPostByID(ctx, post.ID)
		if err != nil {
			return err
		}
		result = pd
		return nil
	})
	return result, err
}
