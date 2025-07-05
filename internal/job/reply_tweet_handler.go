package job

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
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
	scraper              *xscraper.XScraper
	threadURLTemplate    string
	botScreenName        string
}

// NewReplyTweetHandler constructs an ReplyTweetHandler.
func NewReplyTweetHandler(
	logger *slog.Logger,
	mentionService *service.MentionService,
	threadService *service.ThreadService,
	processedMarkService *service.ProcessedMarkService,
	scraper *xscraper.XScraper,
	chromedpCtx ChromedpContext,
	commonConfig *config.CommonConfig,
	botConfig *config.BotConfig,
) *ReplyTweetHandler {
	return &ReplyTweetHandler{
		logger:               logger.With("job_handler", "reply tweet"),
		mentionService:       mentionService,
		threadService:        threadService,
		processedMarkService: processedMarkService,
		chromedpCtx:          chromedpCtx,
		scraper:              scraper,
		threadURLTemplate:    commonConfig.ThreadURLTemplate,
		botScreenName:        botConfig.Username,
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
		return fmt.Errorf("mention not found")
	}

	threadURL := fmt.Sprintf(h.threadURLTemplate, mention.ThreadID)
	replyText := fmt.Sprintf("%s\n\n#threadmirror", threadURL)
	searchQuery := fmt.Sprintf("(%s) (from:%s) filter:replies", threadURL, h.botScreenName)
	// 先检查是否已经回复了该 tweet
	tweets, err := h.scraper.SearchTweets(ctx, searchQuery, 1)
	if err != nil {
		return fmt.Errorf("search tweets %s: %w", searchQuery, err)
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

		// Upload the generated screenshot and obtain the media ID
		uploadRes, err := h.scraper.UploadMedia(ctx, bytes.NewReader(buf), len(buf))
		if err != nil {
			return fmt.Errorf("upload media: %w", err)
		}

		// Construct and send the tweet
		tweet, err := h.scraper.CreateTweet(ctx, xscraper.NewTweet{
			Text:             replyText,
			MediaIDs:         []string{uploadRes.MediaID},
			TaggedUsers:      [][]string{},
			InReplyToTweetId: &payload.MentionID,
		})
		if err != nil {
			return fmt.Errorf("create tweet: %w", err)
		}
		logger.Info("created tweet for mention", "tweet_id", tweet.RestID, "mention_id", payload.MentionID)
		return nil
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
