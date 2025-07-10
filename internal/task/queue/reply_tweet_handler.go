package queue

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

const TypeReplyTweet = "reply_tweet"

type ReplyTweetPayload struct {
	MentionID               string `json:"mention_id"`
	MentionAuthorScreenName string `json:"mention_author_screen_name"`
}

type ChromedpContext context.Context

type ReplyTweetHandler struct {
	logger               *slog.Logger
	mentionService       *service.MentionService
	threadService        *service.ThreadService
	processedMarkService *service.ProcessedMarkService
	chromedpCtx          ChromedpContext
	scrapers             []*xscraper.XScraper
	threadURLTemplate    string
}

// NewReplyTweetHandler constructs an ReplyTweetHandler.
func NewReplyTweetHandler(
	logger *slog.Logger,
	mentionService *service.MentionService,
	threadService *service.ThreadService,
	processedMarkService *service.ProcessedMarkService,
	scrapers []*xscraper.XScraper,
	chromedpCtx ChromedpContext,
	commonConfig *config.CommonConfig,
) *ReplyTweetHandler {
	return &ReplyTweetHandler{
		logger:               logger.With("job_handler", "reply tweet"),
		mentionService:       mentionService,
		threadService:        threadService,
		processedMarkService: processedMarkService,
		chromedpCtx:          chromedpCtx,
		scrapers:             scrapers,
		threadURLTemplate:    commonConfig.ThreadURLTemplate,
	}
}

// NewReplyTweetJob creates a new job for reply tweet for a mention.
func NewReplyTweetJob(mentionID, mentionAuthorScreenName string) (*jobq.Job, error) {
	payload, err := json.Marshal(ReplyTweetPayload{
		MentionID:               mentionID,
		MentionAuthorScreenName: mentionAuthorScreenName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image payload: %w", err)
	}
	return jobq.NewJob(TypeReplyTweet, payload), nil
}

// HandleJob implements the job.JobHandler interface for ReplyTweetHandler.
func (h *ReplyTweetHandler) HandleJob(ctx context.Context, j *jobq.Job) error {
	// sleep 2s to 5s to reduce API burst and mimic human-like interval
	time.Sleep(time.Duration(2+rand.IntN(3)) * time.Second)

	var payload ReplyTweetPayload
	if err := json.Unmarshal(j.Payload, &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}
	if payload.MentionID == "" {
		return fmt.Errorf("mention_id or thread_id is empty")
	}
	logger := h.logger.With("mention_id", payload.MentionID)

	processed, err := h.processedMarkService.IsProcessed(ctx, payload.MentionID, TypeReplyTweet)
	if err != nil {
		return fmt.Errorf("check if thread is processed: %w", err)
	}
	if processed {
		return nil
	}

	mention, err := h.mentionService.GetMentionByID(ctx, payload.MentionID)
	if err != nil {
		if errors.Is(err, errutil.ErrNotFound) {
			return fmt.Errorf("mention not found: %s", payload.MentionID)
		}
		return fmt.Errorf("get mention by id %s: %w", payload.MentionID, err)
	}

	threadURL := fmt.Sprintf(h.threadURLTemplate, mention.ThreadID)
	replyText := fmt.Sprintf("%s\n\n#threadmirror", threadURL)
	searchQuery := fmt.Sprintf("\"%s\" (to:%s) filter:replies", threadURL, payload.MentionAuthorScreenName)

	var tweets []*xscraper.Tweet

	pool := xscraper.NewScraperPool(h.scrapers)
	searchRes, err := xscraper.TryWithResult(pool, func(sc *xscraper.XScraper) ([]*xscraper.Tweet, error) {
		return sc.SearchTweets(ctx, searchQuery, 1)
	})
	if err != nil {
		return fmt.Errorf("no valid scraper found: %w", err)
	}

	tweets = searchRes

	if len(tweets) == 0 {
		thread, err := h.threadService.GetThreadByID(ctx, mention.ThreadID)
		if err != nil {
			return fmt.Errorf("get thread by id %s: %w", mention.ThreadID, err)
		}

		html, err := comm.RenderThread(h.threadURLTemplate, mention.ThreadID, thread, logger)
		if err != nil {
			return fmt.Errorf("render thread id %s: %w", mention.ThreadID, err)
		}

		var buf []byte
		err = chromedp.Run(h.chromedpCtx,
			chromedp.EmulateViewport(485, 0),
			chromedp.Navigate("data:text/html;base64,"+base64.StdEncoding.EncodeToString([]byte(html))),
			chromedp.Sleep(1*time.Second),
			chromedp.FullScreenshot(&buf, 100),
		)
		if err != nil {
			return err
		}

		// Use fallback to upload media and create tweet
		_, err = xscraper.TryWithResult(pool, func(sc *xscraper.XScraper) (*xscraper.Tweet, error) {
			// Upload the generated screenshot and obtain the media ID
			uploadRes, err := sc.UploadMedia(ctx, bytes.NewReader(buf), len(buf))
			if err != nil {
				return nil, err
			}
			return sc.CreateTweet(ctx, xscraper.NewTweet{
				Text:             replyText,
				MediaIDs:         []string{uploadRes.MediaID},
				TaggedUsers:      [][]string{},
				InReplyToTweetId: &payload.MentionID,
			})
		})
		if err != nil {
			return fmt.Errorf("no valid scraper found: %w", err)
		}
		logger.Info("created tweet for mention", "mention_id", payload.MentionID)
	} else {
		logger.Info("tweet already exists", "tweet_id", tweets[0].RestID, "mention_id", payload.MentionID, "search_query", searchQuery)
	}

	err = h.processedMarkService.MarkProcessed(ctx, payload.MentionID, TypeReplyTweet)
	if err != nil {
		return fmt.Errorf("mark thread as processed: %w", err)
	}

	logger.Info("reply tweet for thread", "thread_id", mention.ThreadID)
	return nil
}
