package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/samber/lo"
)

var (
	// Post module error codes: 13000-13999
	ErrCodePost = v1errors.NewErrorCode(v1errors.CheckCode(13000), "Post error")

	// Post operation errors
	ErrCodeFailedToGetPosts = v1errors.NewErrorCode(13001, "failed to get posts")
	ErrCodeFailedToGetPost  = v1errors.NewErrorCode(13003, "failed to get post")

	// Post access errors
	ErrCodePostNotFound = v1errors.NewErrorCode(13006, "post not found")
)

// Post-related methods for V1Handler

// GetPosts handles GET /posts
func (h *V1Handler) GetPosts(c *gin.Context, params GetPostsParams) {
	currentUserID := auth.CurrentUserID(c)

	limit, offset := ExtractPaginationParams(&params)

	// Get posts
	posts, total, err := h.postService.GetPosts(c.Request.Context(), currentUserID, limit, offset)
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetPosts))
		return
	}

	apiPosts := lo.Map(posts, func(post service.PostSummaryDetail, _ int) PostSummary {
		return h.convertPostSummaryToAPI(post)
	})

	PaginatedJSON(c, apiPosts, total, limit, offset)
}

// GetPostsId handles GET /posts/{id}
func (h *V1Handler) GetPostsId(c *gin.Context, id string) {
	thread, err := h.postService.GetPostByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodePostNotFound))
			return
		}
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetPost))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": h.convertPostDetailToAPI(*thread),
	})
}

// Post Helper functions

// Conversion functions from service types to API types

func (h *V1Handler) convertPostSummaryToAPI(post service.PostSummaryDetail) PostSummary {
	var author *PostAuthor
	if post.Author != nil {
		author = &PostAuthor{
			Id:              post.Author.ID,
			Name:            post.Author.Name,
			ScreenName:      post.Author.ScreenName,
			ProfileImageUrl: post.Author.ProfileImageURL,
		}
	}

	return PostSummary{
		Id:             post.ID,
		ContentPreview: post.ContentPreview,
		Author:         author,
		CreatedAt:      post.CreatedAt,
		NumTweets:      post.NumTweets,
	}
}

func (h *V1Handler) convertPostDetailToAPI(post service.PostDetail) PostDetail {
	var author *PostAuthor
	if post.Author != nil {
		author = &PostAuthor{
			Id:              post.Author.ID,
			Name:            post.Author.Name,
			ScreenName:      post.Author.ScreenName,
			ProfileImageUrl: post.Author.ProfileImageURL,
		}
	}

	var apiThreads []Tweet
	if len(post.Threads) > 0 {
		apiThreads = lo.Map(post.Threads, func(tweet *xscraper.Tweet, _ int) Tweet {
			return h.convertXScraperTweetToAPI(tweet)
		})
	}

	return PostDetail{
		Id:        post.ID,
		Author:    author,
		CreatedAt: post.CreatedAt,
		Threads:   &apiThreads,
	}
}

// convertXScraperTweetToAPI converts xscraper.Tweet to API Tweet type
func (h *V1Handler) convertXScraperTweetToAPI(tweet *xscraper.Tweet) Tweet {
	if tweet == nil {
		return Tweet{}
	}

	// Convert author
	var author *TweetUser
	if tweet.Author != nil {
		author = &TweetUser{
			Id:              tweet.Author.ID,
			RestId:          tweet.Author.RestID,
			Name:            tweet.Author.Name,
			ScreenName:      tweet.Author.ScreenName,
			ProfileImageUrl: tweet.Author.ProfileImageURL,
			Description:     tweet.Author.Description,
			FollowersCount:  tweet.Author.FollowersCount,
			FriendsCount:    tweet.Author.FriendsCount,
			StatusesCount:   tweet.Author.StatusesCount,
			CreatedAt:       tweet.Author.CreatedAt,
			Verified:        tweet.Author.Verified,
			IsBlueVerified:  tweet.Author.IsBlueVerified,
		}
	}

	// Convert entities
	entities := &TweetEntities{
		Hashtags:     tweet.Entities.Hashtags,
		Symbols:      tweet.Entities.Symbols,
		Urls:         tweet.Entities.URLs,
		UserMentions: tweet.Entities.UserMentions,
	}

	// Convert media
	if len(tweet.Entities.Media) > 0 {
		media := lo.Map(tweet.Entities.Media, func(m xscraper.MediaInfo, _ int) MediaInfo {
			return MediaInfo{
				Id:          m.ID,
				MediaKey:    m.MediaKey,
				Type:        m.Type,
				Url:         m.URL,
				DisplayUrl:  m.DisplayURL,
				ExpandedUrl: m.ExpandedURL,
				AltText:     &m.AltText,
				Width:       &m.Width,
				Height:      &m.Height,
			}
		})
		entities.Media = &media
	}

	// Convert stats
	stats := TweetStats{
		ReplyCount:    tweet.Stats.ReplyCount,
		RetweetCount:  tweet.Stats.RetweetCount,
		FavoriteCount: tweet.Stats.FavoriteCount,
		QuoteCount:    tweet.Stats.QuoteCount,
		BookmarkCount: tweet.Stats.BookmarkCount,
	}
	if tweet.Stats.ViewCount > 0 {
		stats.ViewCount = &tweet.Stats.ViewCount
	}

	// Convert quoted tweet if exists
	var quotedTweet *Tweet
	if tweet.QuotedTweet != nil {
		quoted := h.convertXScraperTweetToAPI(tweet.QuotedTweet)
		quotedTweet = &quoted
	}

	return Tweet{
		Id:                tweet.ID,
		RestId:            tweet.RestID,
		Text:              tweet.Text,
		CreatedAt:         tweet.CreatedAt,
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
	}
}
