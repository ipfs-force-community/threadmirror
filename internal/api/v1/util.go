package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
	"github.com/samber/lo"
)

type PaginationParm interface {
	GetLimit() *PageLimit
	GetOffset() *PageOffset
}

func ExtractPaginationParams(p PaginationParm) (limit int, offset int) {
	limit, offset = 20, 0
	if p.GetLimit() != nil {
		limit = *p.GetLimit()
	}

	if p.GetOffset() != nil {
		offset = *p.GetOffset()
	}

	return
}

func PaginatedJSON(c *gin.Context, data any, total int64, limit, offset int) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": struct {
			Total  int64 `json:"total"`
			Limit  int   `json:"limit"`
			Offset int   `json:"offset"`
		}{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func ParseStringUUID(c *gin.Context, id string, errCode v1errors.ErrorCode) (string, bool) {
	if id == "" {
		_ = c.Error(v1errors.NotFound(nil).WithCode(errCode))
		return "", false
	}
	return id, true
}

// safeStringFromMap safely extracts a string value from a map with a fallback default
func safeStringFromMap(m map[string]interface{}, key string, defaultVal ...string) string {
	if m == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}

	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// safeIntSliceFromInterface safely converts interface{} to []int
func safeIntSliceFromInterface(v interface{}) []int {
	if v == nil {
		return []int{}
	}

	arr, ok := v.([]interface{})
	if !ok {
		return []int{}
	}

	res := make([]int, 0, len(arr))
	for _, x := range arr {
		switch val := x.(type) {
		case int:
			res = append(res, val)
		case float64:
			res = append(res, int(val))
		case int64:
			res = append(res, int(val))
		case float32:
			res = append(res, int(val))
		default:
			// Skip invalid values instead of returning nil
			continue
		}
	}
	return res
}

// convertMapToHashtag safely converts a map to Hashtag struct
func convertMapToHashtag(m map[string]interface{}) Hashtag {
	return Hashtag{
		Text:    safeStringFromMap(m, "text"),
		Indices: safeIntSliceFromInterface(m["indices"]),
	}
}

// convertMapToSymbol safely converts a map to Symbol struct
func convertMapToSymbol(m map[string]interface{}) Symbol {
	return Symbol{
		Text:    safeStringFromMap(m, "text"),
		Indices: safeIntSliceFromInterface(m["indices"]),
	}
}

// convertMapToUserMention safely converts a map to UserMention struct
func convertMapToUserMention(m map[string]interface{}) UserMention {
	return UserMention{
		Id:         safeStringFromMap(m, "id_str"),
		ScreenName: safeStringFromMap(m, "screen_name"),
		Name:       safeStringFromMap(m, "name"),
		Indices:    safeIntSliceFromInterface(m["indices"]),
	}
}

// convertTweetAuthor safely converts xscraper User to API TweetUser
func convertTweetAuthor(author *xscraper.User) *TweetUser {
	if author == nil {
		return nil
	}

	return &TweetUser{
		Id:              author.ID,
		RestId:          author.RestID,
		Name:            author.Name,
		ScreenName:      author.ScreenName,
		ProfileImageUrl: author.ProfileImageURL,
		Description:     author.Description,
		FollowersCount:  author.FollowersCount,
		FriendsCount:    author.FriendsCount,
		StatusesCount:   author.StatusesCount,
		CreatedAt:       author.CreatedAt,
		Verified:        author.Verified,
		IsBlueVerified:  author.IsBlueVerified,
	}
}

// convertTweetEntities safely converts generated.Entities to API TweetEntities
func convertTweetEntities(entities *generated.Entities) *TweetEntities {
	if entities == nil {
		return &TweetEntities{
			Hashtags:     []Hashtag{},
			Symbols:      []Symbol{},
			Urls:         []Url{},
			UserMentions: []UserMention{},
		}
	}

	apiEntities := &TweetEntities{
		Hashtags:     []Hashtag{},
		Symbols:      []Symbol{},
		Urls:         []Url{},
		UserMentions: []UserMention{},
	}

	// Convert hashtags
	if entities.Hashtags != nil {
		apiEntities.Hashtags = lo.Map(entities.Hashtags, func(h generated.Hashtag, _ int) Hashtag {
			return convertMapToHashtag(h)
		})
	}

	// Convert symbols
	if entities.Symbols != nil {
		apiEntities.Symbols = lo.Map(entities.Symbols, func(s generated.Symbol, _ int) Symbol {
			return convertMapToSymbol(s)
		})
	}

	// Convert URLs
	if entities.Urls != nil {
		apiEntities.Urls = lo.Map(entities.Urls, func(u generated.Url, _ int) Url {
			return Url{
				DisplayUrl:  u.DisplayUrl,
				ExpandedUrl: u.ExpandedUrl,
				Indices:     u.Indices,
				Url:         u.Url,
			}
		})
	}

	// Convert user mentions
	if entities.UserMentions != nil {
		apiEntities.UserMentions = lo.Map(entities.UserMentions, func(m generated.UserMention, _ int) UserMention {
			return convertMapToUserMention(m)
		})
	}

	// Convert media
	if entities.Media != nil && len(*entities.Media) > 0 {
		media := lo.Map(*entities.Media, func(m generated.Media, _ int) Media {
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
		apiEntities.Media = &media
	}

	return apiEntities
}

// convertTweetStats safely converts xscraper TweetStats to API TweetStats
func convertTweetStats(stats xscraper.TweetStats) TweetStats {
	apiStats := TweetStats{
		ReplyCount:    stats.ReplyCount,
		RetweetCount:  stats.RetweetCount,
		FavoriteCount: stats.FavoriteCount,
		QuoteCount:    stats.QuoteCount,
		BookmarkCount: stats.BookmarkCount,
	}

	if stats.ViewCount > 0 {
		apiStats.ViewCount = &stats.ViewCount
	}

	return apiStats
}

// convertTweetRichText safely converts generated rich text to API rich text
func convertTweetRichText(richText *generated.NoteTweetResultRichText) NoteTweetRichText {
	apiRichtext := NoteTweetRichText{
		RichtextTags: []NoteTweetRichTextTag{},
	}

	if richText != nil && richText.RichtextTags != nil {
		apiRichtext.RichtextTags = lo.Map(richText.RichtextTags, func(tag generated.NoteTweetResultRichTextTag, _ int) NoteTweetRichTextTag {
			return NoteTweetRichTextTag{
				FromIndex: tag.FromIndex,
				ToIndex:   tag.ToIndex,
				RichtextTypes: lo.Map(tag.RichtextTypes, func(t generated.NoteTweetResultRichTextTagRichtextTypes, _ int) NoteTweetRichTextTagRichtextTypes {
					return NoteTweetRichTextTagRichtextTypes(t)
				}),
			}
		})
	}

	return apiRichtext
}

// Common error handling functions to reduce duplicate code

// HandleBadRequestError is a helper for bad request errors
func HandleBadRequestError(c *gin.Context, err error) {
	_ = c.Error(v1errors.BadRequest(err).WithCode(v1errors.ErrCodeBadRequest))
}

// HandleInternalServerError is a helper for internal server errors
func HandleInternalServerError(c *gin.Context, err error) {
	_ = c.Error(v1errors.InternalServerError(err).WithCode(v1errors.ErrCodeInternalError))
}

// HandleNotFoundError is a helper for not found errors
func HandleNotFoundError(c *gin.Context, err error) {
	_ = c.Error(v1errors.NotFound(err).WithCode(v1errors.ErrCodeNotFound))
}
