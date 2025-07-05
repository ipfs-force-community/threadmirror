package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/ipfs"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs/go-cid"
)

type ThreadDetail struct {
	ID             string            `json:"id"`
	CID            string            `json:"cid"`
	ContentPreview string            `json:"content_preview"`
	NumTweets      int               `json:"numTweets"`
	Tweets         []*xscraper.Tweet `json:"tweets,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type ThreadRepoInterface interface {
	GetThreadByID(ctx context.Context, id string) (*model.Thread, error)
	CreateThread(ctx context.Context, thread *model.Thread) error
	GetTweetsByIDs(ctx context.Context, ids []string) (map[string]*model.Thread, error)
}

type ThreadService struct {
	threadRepo ThreadRepoInterface
	storage    ipfs.Storage
}

func NewThreadService(threadRepo ThreadRepoInterface, storage ipfs.Storage) *ThreadService {
	return &ThreadService{threadRepo: threadRepo, storage: storage}
}

func (s *ThreadService) GetThreadByID(ctx context.Context, id string) (*ThreadDetail, error) {
	thread, err := s.threadRepo.GetThreadByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get thread: %w", err)
	}
	if thread == nil {
		return nil, nil
	}

	// Load tweets from IPFS
	tweets, err := s.loadTweetsFromIPFS(context.Background(), thread.CID)
	if err != nil {
		return nil, fmt.Errorf("load from ipfs %s: %w", thread.CID, err)
	}

	return &ThreadDetail{
		ID:             thread.ID,
		CID:            thread.CID,
		ContentPreview: thread.Summary,
		NumTweets:      thread.NumTweets,
		Tweets:         tweets,
		CreatedAt:      thread.CreatedAt,
	}, nil

}

// loadTweetsFromIPFS loads tweets from IPFS using the CID
func (s *ThreadService) loadTweetsFromIPFS(ctx context.Context, cidStr string) ([]*xscraper.Tweet, error) {
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
