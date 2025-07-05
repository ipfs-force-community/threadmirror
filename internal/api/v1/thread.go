package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
	"github.com/samber/lo"
)

// Thread access errors
var ErrCodeThreadNotFound = v1errors.NewErrorCode(14001, "thread not found")

// GetMentionsId handles GET /mentions/{id}
func (h *V1Handler) GetThreadId(c *gin.Context, id string) {
	thread, err := h.threadService.GetThreadByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetMention))
		return
	}
	if thread == nil {
		_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodeThreadNotFound))
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

	return ThreadDetail{
		Id:             thread.ID,
		Cid:            thread.CID,
		ContentPreview: thread.ContentPreview,
		NumTweets:      thread.NumTweets,
		CreatedAt:      thread.CreatedAt,
		Tweets:         &apiTweets,
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
		Hashtags: lo.Map(tweet.Entities.Hashtags, func(h generated.Hashtag, _ int) Hashtag {
			return Hashtag{
				Text:    h["text"].(string),
				Indices: toIntSlice(h["indices"]),
			}
		}),
		Symbols: lo.Map(tweet.Entities.Symbols, func(s generated.Symbol, _ int) Symbol {
			return Symbol{
				Text:    s["text"].(string),
				Indices: toIntSlice(s["indices"]),
			}
		}),
		Urls: lo.Map(tweet.Entities.Urls, func(u generated.Url, _ int) Url {
			return Url{
				DisplayUrl:  u.DisplayUrl,
				ExpandedUrl: u.ExpandedUrl,
				Indices:     u.Indices,
				Url:         u.Url,
			}
		}),
		UserMentions: lo.Map(tweet.Entities.UserMentions, func(m generated.UserMention, _ int) UserMention {
			return UserMention{
				Id:         m["id_str"].(string),
				ScreenName: m["screen_name"].(string),
				Name:       m["name"].(string),
				Indices:    toIntSlice(m["indices"]),
			}
		}),
	}

	// Convert media
	if tweet.Entities.Media != nil && len(*tweet.Entities.Media) > 0 {
		media := lo.Map(*tweet.Entities.Media, func(m generated.Media, _ int) Media {
			return Media{
				IdStr:         m.IdStr,
				MediaKey:      m.MediaKey,
				Type:          string(m.Type),
				Url:           m.Url,
				DisplayUrl:    m.DisplayUrl,
				ExpandedUrl:   m.ExpandedUrl,
				MediaUrlHttps: m.MediaUrlHttps,
				Indices:       m.Indices,
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

// 工具函数：interface{} 转 []int
func toIntSlice(v interface{}) []int {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	res := make([]int, 0, len(arr))
	for _, x := range arr {
		if i, ok := x.(int); ok {
			res = append(res, i)
		}
	}
	return res
}
