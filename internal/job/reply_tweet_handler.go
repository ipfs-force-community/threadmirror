package job

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
)

const TypeReplyTweet = "reply_tweet"

type ReplyTweetPayload struct {
	MentionID string `json:"mention_id"`
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
func NewReplyTweetJob(mentionID string) (*jobq.Job, error) {
	payload, err := json.Marshal(ReplyTweetPayload{MentionID: mentionID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal image payload: %w", err)
	}
	return jobq.NewJob(TypeReplyTweet, payload), nil
}

// HandleJob implements the job.JobHandler interface for ReplyTweetHandler.
func (h *ReplyTweetHandler) HandleJob(ctx context.Context, j *jobq.Job) error {
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
		return fmt.Errorf("get mention by id %s: %w", payload.MentionID, err)
	}
	if mention == nil {
		return fmt.Errorf("mention not found: %s", payload.MentionID)
	}

	threadURL := fmt.Sprintf(h.threadURLTemplate, mention.ThreadID)
	replyText := fmt.Sprintf("%s\n\n#threadmirror", threadURL)
	searchQuery := fmt.Sprintf("(%s) filter:replies", threadURL)
	// 先检查是否已经回复了该 tweet
	var tweets []*xscraper.Tweet

	scrapers := make([]*xscraper.XScraper, len(h.scrapers))
	copy(scrapers, h.scrapers)

	rand.Shuffle(len(scrapers), func(i, j int) {
		scrapers[i], scrapers[j] = scrapers[j], scrapers[i]
	})

	valid := false
	for _, scraper := range scrapers {
		tweets, err = scraper.SearchTweets(ctx, searchQuery, 1)
		if err != nil {
			logger.Error("search tweets", "error", err, "search_query", searchQuery)
			continue
		}
		valid = true
		break
	}

	if !valid {
		return fmt.Errorf("no valid scraper found")
	}

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

		valid = false
		for _, scraper := range h.scrapers {
			// Upload the generated screenshot and obtain the media ID
			uploadRes, err := scraper.UploadMedia(ctx, bytes.NewReader(buf), len(buf))
			if err != nil {
				logger.Error("upload media", "error", err)
				continue
			}

			// Construct and send the tweet
			tweet, err := scraper.CreateTweet(ctx, xscraper.NewTweet{
				Text:             replyText,
				MediaIDs:         []string{uploadRes.MediaID},
				TaggedUsers:      [][]string{},
				InReplyToTweetId: &payload.MentionID,
			})
			if err != nil {
				logger.Error("create tweet", "error", err)
				continue
			}
			logger.Info("created tweet for mention", "tweet_id", tweet.RestID, "mention_id", payload.MentionID)
			valid = true
			break
		}

		if !valid {
			return fmt.Errorf("no valid scraper found")
		}
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
