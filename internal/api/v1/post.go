package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
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
	posts, total, err := h.postService.GetPosts(currentUserID, limit, offset)
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetPosts))
		return
	}

	// Convert to API response format using lo.Map
	apiPosts := lo.Map(posts, func(post service.PostSummaryDetail, _ int) Post {
		return h.convertPostSummaryToAPI(post)
	})

	PaginatedJSON(c, apiPosts, total, limit, offset)
}

// GetPostsId handles GET /posts/{id}
func (h *V1Handler) GetPostsId(c *gin.Context, id string) {
	authInfo := auth.MustAuthInfo(c)

	postID, ok := ParseStringUUID(c, id, ErrCodePostNotFound)
	if !ok {
		return
	}

	post, err := h.postService.GetPostByID(postID, authInfo.UserID)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodePostNotFound))
			return
		}
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetPost))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": h.convertPostDetailToAPI(*post),
	})
}

// Post Helper functions

// Conversion functions from service types to API types

func (h *V1Handler) convertPostSummaryToAPI(post service.PostSummaryDetail) Post {
	return Post{
		Id:      post.ID,
		Content: post.ContentPreview,
		User:    h.convertUserProfileSummaryToAPI(post.User),
		Images: []struct {
			ImageId string `json:"image_id"`
		}{}, // Summary doesn't include all images
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.CreatedAt, // Use created_at as fallback
	}
}

func (h *V1Handler) convertPostDetailToAPI(post service.PostDetail) PostDetails {
	images := lo.Map(post.Images, func(img service.PostImageDetail, _ int) struct {
		ImageId string `json:"image_id"`
	} {
		return struct {
			ImageId string `json:"image_id"`
		}{
			ImageId: img.ImageID,
		}
	})

	return PostDetails{
		Id:        post.ID,
		Content:   post.Content,
		User:      h.convertUserProfileSummaryToAPI(post.User),
		Images:    images,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}

func (h *V1Handler) convertUserProfileSummaryToAPI(
	user service.UserProfileSummary,
) UserProfileSummary {
	return UserProfileSummary{
		UserId:    user.UserID,
		DisplayId: user.DisplayID,
		Nickname:  user.Nickname,
		Bio:       user.Bio,
	}
}
