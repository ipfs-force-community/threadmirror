package v1

import (
	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	"github.com/samber/lo"
)

var (
	// Mention module error codes: 13000-13999
	ErrCodeMention = v1errors.NewErrorCode(v1errors.CheckCode(13000), "Mention error")

	// Mention operation errors
	ErrCodeFailedToGetMentions = v1errors.NewErrorCode(13001, "failed to get mentions")
	ErrCodeFailedToGetMention  = v1errors.NewErrorCode(13003, "failed to get mention")
)

// Mention-related methods for V1Handler

// GetMentions handles GET /mentions
func (h *V1Handler) GetMentions(c *gin.Context, params GetMentionsParams) {
	currentUserID := auth.CurrentUserID(c)

	limit, offset := ExtractPaginationParams(&params)

	// Get mentions
	mentions, total, err := h.mentionService.GetMentions(c.Request.Context(), currentUserID, limit, offset)
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetMentions))
		return
	}

	apiMentions := lo.Map(mentions, func(mention service.MentionSummary, _ int) MentionSummary {
		return h.convertMentionSummaryToAPI(mention)
	})

	PaginatedJSON(c, apiMentions, total, limit, offset)
}

// Mention Helper functions

// Conversion functions from service types to API types

func (h *V1Handler) convertMentionSummaryToAPI(mention service.MentionSummary) MentionSummary {
	var threadAuthor *ThreadAuthor
	if mention.ThreadAuthor != nil {
		threadAuthor = &ThreadAuthor{
			Id:              mention.ThreadAuthor.ID,
			Name:            mention.ThreadAuthor.Name,
			ScreenName:      mention.ThreadAuthor.ScreenName,
			ProfileImageUrl: mention.ThreadAuthor.ProfileImageURL,
		}
	}

	return MentionSummary{
		Id:              mention.ID,
		Cid:             mention.CID,
		ContentPreview:  mention.ContentPreview,
		ThreadId:        mention.ThreadID,
		ThreadAuthor:    threadAuthor,
		CreatedAt:       mention.CreatedAt,
		MentionCreateAt: mention.MentionCreateAt,
		NumTweets:       mention.NumTweets,
		Status:          MentionSummaryStatus(mention.Status),
	}
}
