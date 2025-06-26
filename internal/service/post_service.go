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
	Threads   []*xscraper.Tweet `json:"threads,omitempty"`
	Author    *PostAuthor       `json:"author"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// PostSummaryDetail represents a post summary for list views
type PostSummaryDetail struct {
	ID             string      `json:"id"`
	ContentPreview string      `json:"content_preview"`
	Author         *PostAuthor `json:"author"`
	CreatedAt      time.Time   `json:"created_at"`
}

// CreatePostRequest represents a request to create a new post
type CreatePostRequest struct {
	Tweets []*xscraper.Tweet
}

// PostRepoInterface defines the interface for post repo operations
type PostRepoInterface interface {
	// Post CRUD
	GetPostByID(id string) (*model.Post, error)
	CreatePost(post *model.Post) error
	GetPosts(
		userID string,
		limit, offset int,
	) ([]model.Post, int64, error)
	GetPostsByUser(userID string, limit, offset int) ([]model.Post, int64, error)
}

// PostService provides business logic for post operations
type PostService struct {
	postRepo PostRepoInterface
	llm      llm.Model
	storage  ipfs.Storage
}

// NewPostService creates a new post service
func NewPostService(
	postRepo PostRepoInterface,
	llm llm.Model,
	storage ipfs.Storage,
) *PostService {
	return &PostService{
		postRepo: postRepo,
		llm:      llm,
		storage:  storage,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(
	ctx context.Context,
	userID string,
	req *CreatePostRequest,
) (*PostDetail, error) {

	jsonTweets, err := json.Marshal(req.Tweets)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tweets: %w", err)
	}

	cid, err := s.storage.Add(ctx, bytes.NewReader(jsonTweets))
	if err != nil {
		return nil, fmt.Errorf("failed to add tweets to IPFS: %w", err)
	}

	// Generate AI summary of the tweets
	summary, err := s.generateTweetsSummary(ctx, req.Tweets)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI summary: %w", err)
	}

	// Extract author information from the first tweet (main tweet)
	var authorID, authorName, authorScreenName, authorProfileImageURL string
	if len(req.Tweets) > 0 && req.Tweets[0].Author != nil {
		author := req.Tweets[0].Author
		authorID = author.ID
		authorName = author.Name
		authorScreenName = author.ScreenName
		authorProfileImageURL = author.ProfileImageURL
	}

	// Create post record
	post := &model.Post{
		UserID:                userID,
		AuthorID:              authorID,
		CID:                   cid.String(),
		Summary:               summary,
		AuthorName:            authorName,
		AuthorScreenName:      authorScreenName,
		AuthorProfileImageURL: authorProfileImageURL,
	}

	if err := s.postRepo.CreatePost(post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Return the created post details
	return s.GetPostByID(ctx, post.ID)
}

// GetPostByID retrieves a post by ID
func (s *PostService) GetPostByID(ctx context.Context, postID string) (*PostDetail, error) {
	post, err := s.postRepo.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return s.buildPostDetail(ctx, post)
}

// GetPosts retrieves posts based on feed type
func (s *PostService) GetPosts(
	userID string,
	limit, offset int,
) ([]PostSummaryDetail, int64, error) {
	posts, total, err := s.postRepo.GetPosts(userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts: %w", err)
	}

	postSummaries := make([]PostSummaryDetail, 0, len(posts))
	for _, post := range posts {
		postSummaries = append(postSummaries, *s.buildPostSummary(&post))
	}

	return postSummaries, total, nil
}

// buildPostDetail builds a PostDetail from a model.Post
func (s *PostService) buildPostDetail(
	ctx context.Context,
	post *model.Post,
) (*PostDetail, error) {
	// Build author information
	var author *PostAuthor
	if post.AuthorID != "" {
		author = &PostAuthor{
			ID:              post.AuthorID,
			Name:            post.AuthorName,
			ScreenName:      post.AuthorScreenName,
			ProfileImageURL: post.AuthorProfileImageURL,
		}
	}

	// Load threads from IPFS
	threads, _ := s.loadThreadsFromIPFS(ctx, post.CID)

	return &PostDetail{
		ID:        post.ID,
		Threads:   threads,
		Author:    author,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}, nil
}

// buildPostSummary builds a PostSummaryDetail from a model.Post
func (s *PostService) buildPostSummary(
	post *model.Post,
) *PostSummaryDetail {
	// Build author information
	var author *PostAuthor
	if post.AuthorID != "" {
		author = &PostAuthor{
			ID:              post.AuthorID,
			Name:            post.AuthorName,
			ScreenName:      post.AuthorScreenName,
			ProfileImageURL: post.AuthorProfileImageURL,
		}
	}

	return &PostSummaryDetail{
		ID:             post.ID,
		ContentPreview: post.Summary, // Use summary as content preview
		Author:         author,
		CreatedAt:      post.CreatedAt,
	}
}

// generateTweetsSummary generates AI summary for tweets
func (s *PostService) generateTweetsSummary(ctx context.Context, tweets []*xscraper.Tweet) (string, error) {
	// Extract tweet content for summarization
	var tweetTexts []string
	for _, tweet := range tweets {
		if tweet.Text != "" {
			tweetTexts = append(tweetTexts, tweet.Text)
		}
	}

	if len(tweetTexts) == 0 {
		return "No content to summarize", nil
	}

	// Combine all tweet texts
	allTweetText := strings.Join(tweetTexts, "\n\n")

	// Create prompt for AI summarization
	prompt := fmt.Sprintf(`Please provide a concise summary (maximum 200 characters) of the following tweets in Chinese:

%s

Summary:`, allTweetText)

	// Generate summary using LLM
	summary, err := llms.GenerateFromSinglePrompt(ctx, s.llm, prompt,
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	// Ensure summary is not too long
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	return strings.TrimSpace(summary), nil
}

// loadThreadsFromIPFS loads threads from IPFS using the CID
func (s *PostService) loadThreadsFromIPFS(ctx context.Context, cidStr string) ([]*xscraper.Tweet, error) {
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
	defer reader.Close()

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
