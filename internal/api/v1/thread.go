package v1

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/internal/task/queue"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/samber/lo"
)

// Thread access errors
var ErrCodeThreadNotFound = v1errors.NewErrorCode(14001, "thread not found")

// GetMentionsId handles GET /mentions/{id}
func (h *V1Handler) GetThreadId(c *gin.Context, id string) {
	thread, err := h.threadService.GetThreadByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, errutil.ErrNotFound) {
			_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodeThreadNotFound))
			return
		}
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetMention))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": h.convertThreadDetailToAPI(thread),
	})
}

func (h *V1Handler) convertThreadDetailToAPI(thread *service.ThreadDetail) ThreadDetail {
	var apiTweets []Tweet
	if len(thread.Tweets) > 0 {
		apiTweets = lo.Map(thread.Tweets, func(tweet *xscraper.Tweet, _ int) Tweet {
			return h.convertXScraperTweetToAPI(tweet)
		})
	}

	// Convert author if available
	var apiAuthor *ThreadAuthor
	if thread.Author != nil {
		apiAuthor = &ThreadAuthor{
			Id:              thread.Author.ID,
			Name:            thread.Author.Name,
			ScreenName:      thread.Author.ScreenName,
			ProfileImageUrl: thread.Author.ProfileImageURL,
		}
	}

	var status ThreadDetailStatus
	if thread.RetryCount >= h.commonConfig.ThreadMaxRetries-1 {
		status = ThreadDetailStatusFailed
	} else {
		status = ThreadDetailStatus(thread.Status)
	}

	return ThreadDetail{
		Id:             thread.ID,
		Cid:            thread.CID,
		ContentPreview: thread.ContentPreview,
		NumTweets:      thread.NumTweets,
		CreatedAt:      thread.CreatedAt,
		Tweets:         &apiTweets,
		Status:         status,
		Author:         apiAuthor,
	}
}

// convertXScraperTweetToAPI converts xscraper.Tweet to API Tweet type
func (h *V1Handler) convertXScraperTweetToAPI(tweet *xscraper.Tweet) Tweet {
	if tweet == nil {
		return Tweet{}
	}

	// Convert components using dedicated converter functions
	author := convertTweetAuthor(tweet.Author)
	entities := convertTweetEntities(&tweet.Entities)
	stats := convertTweetStats(tweet.Stats)
	richtext := convertTweetRichText(tweet.RichText)

	// Convert quoted tweet recursively if exists
	var quotedTweet *Tweet
	if tweet.QuotedTweet != nil {
		quoted := h.convertXScraperTweetToAPI(tweet.QuotedTweet)
		quotedTweet = &quoted
	}

	// Convert DisplayTextRange if exists
	var displayTextRange *[]int
	if len(tweet.DisplayTextRange) > 0 {
		displayTextRange = &tweet.DisplayTextRange
	}

	return Tweet{
		Id:                tweet.ID,
		RestId:            tweet.RestID,
		Text:              tweet.Text,
		CreatedAt:         tweet.CreatedAt,
		DisplayTextRange:  displayTextRange,
		Author:            author,
		Entities:          entities,
		Stats:             stats,
		IsRetweet:         tweet.IsRetweet,
		IsReply:           tweet.IsReply,
		IsQuoteStatus:     tweet.IsQuoteStatus,
		ConversationId:    tweet.ConversationID,
		InReplyToStatusId: &tweet.InReplyToStatusID,
		InReplyToUserId:   &tweet.InReplyToUserID,
		QuotedTweet:       quotedTweet,
		HasBirdwatchNotes: tweet.HasBirdwatchNotes,
		Lang:              tweet.Lang,
		Source:            &tweet.Source,
		PossiblySensitive: tweet.PossiblySensitive,
		IsTranslatable:    tweet.IsTranslatable,
		Views:             &tweet.Views,
		IsNoteTweet:       tweet.IsNoteTweet,
		Richtext:          richtext,
	}
}

// PostThreadScrape handles POST /thread/scrape
func (h *V1Handler) PostThreadScrape(c *gin.Context) {
	var req PostThreadScrapeJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleBadRequestError(c, err)
		return
	}

	// Get current user ID
	currentUserID := auth.CurrentUserID(c)
	if currentUserID == "" {
		_ = c.Error(v1errors.Forbidden(fmt.Errorf("user not authenticated")))
		return
	}

	// Parse Twitter URL to extract tweet ID
	tweetID, err := util.ExtractTweetID(req.Url)
	if err != nil {
		HandleBadRequestError(c, err)
		return
	}

	// Check if thread already exists
	existingThread, err := h.threadService.GetThreadByID(c.Request.Context(), tweetID)
	if err != nil && !errors.Is(err, errutil.ErrNotFound) {
		HandleInternalServerError(c, err)
		return
	}

	// If thread already exists, return it with 409 status
	if existingThread != nil {
		c.JSON(http.StatusConflict, gin.H{
			"data":    h.convertThreadDetailToAPI(existingThread),
			"message": "Thread already exists",
		})
		return
	}

	// Create mention record and pending thread
	_, err = h.mentionService.CreateMention(c.Request.Context(), currentUserID, tweetID, nil, time.Now())
	if err != nil {
		HandleInternalServerError(c, err)
		return
	}

	// Create and enqueue thread scrape job
	job, err := queue.NewThreadScrapeJob(tweetID)
	if err != nil {
		HandleInternalServerError(c, err)
		return
	}

	jobID, err := h.jobQueueClient.Enqueue(c.Request.Context(), job)
	if err != nil {
		HandleInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":   jobID,
		"tweet_id": tweetID,
		"message":  "Thread scraping job has been queued and mention created",
	})
}
