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
	ErrCodeFailedToGetPosts   = v1errors.NewErrorCode(13001, "failed to get posts")
	ErrCodeFailedToCreatePost = v1errors.NewErrorCode(13002, "failed to create post")
	ErrCodeFailedToGetPost    = v1errors.NewErrorCode(13003, "failed to get post")
	ErrCodeFailedToUpdatePost = v1errors.NewErrorCode(13004, "failed to update post")
	ErrCodeFailedToDeletePost = v1errors.NewErrorCode(13005, "failed to delete post")

	// Post access errors
	ErrCodePostNotFound  = v1errors.NewErrorCode(13006, "post not found")
	ErrCodeNotPostAuthor = v1errors.NewErrorCode(13007, "not post author")
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

// PostPosts handles POST /posts
func (h *V1Handler) PostPosts(c *gin.Context) {
	authInfo := auth.MustAuthInfo(c)

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(v1errors.InvalidRequestBody(err))
		return
	}

	// Convert to service request
	var imageIDs []string
	if req.ImageIds != nil {
		imageIDs = *req.ImageIds
	}

	serviceReq := &service.CreatePostRequest{
		Content:  req.Content,
		ImageIDs: imageIDs,
	}

	// Create post
	post, err := h.postService.CreatePost(authInfo.UserID, serviceReq)
	if err != nil {
		_ = c.Error(v1errors.BadRequest(err).WithCode(ErrCodeFailedToCreatePost))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": h.convertPostDetailToAPI(*post),
	})
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

// PutPostsId handles PUT /posts/{id}
func (h *V1Handler) PutPostsId(c *gin.Context, id string) {
	authInfo := auth.MustAuthInfo(c)

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(v1errors.InvalidRequestBody(err))
		return
	}

	// Convert to service request
	var imageIDs []string
	if req.ImageIds != nil {
		imageIDs = *req.ImageIds
	}

	serviceReq := &service.UpdatePostRequest{
		Content:  req.Content,
		ImageIDs: imageIDs,
	}
	postID, ok := ParseStringUUID(c, id, ErrCodePostNotFound)
	if !ok {
		return
	}

	// Update post
	post, err := h.postService.UpdatePost(postID, authInfo.UserID, serviceReq)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodePostNotFound))
			return
		case errors.Is(err, service.ErrUnauthorized):
			_ = c.Error(v1errors.Forbidden(err).WithCode(ErrCodeNotPostAuthor))
			return
		default:
			_ = c.Error(v1errors.BadRequest(err).WithCode(ErrCodeFailedToUpdatePost))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": h.convertPostDetailToAPI(*post),
	})
}

// DeletePostsId handles DELETE /posts/{id}
func (h *V1Handler) DeletePostsId(c *gin.Context, id string) {
	authInfo := auth.MustAuthInfo(c)

	postID, ok := ParseStringUUID(c, id, ErrCodePostNotFound)
	if !ok {
		return
	}

	err := h.postService.DeletePost(postID, authInfo.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodePostNotFound))
			return
		case errors.Is(err, service.ErrUnauthorized):
			_ = c.Error(v1errors.Forbidden(err).WithCode(ErrCodeNotPostAuthor))
			return
		default:
			_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToDeletePost))
			return
		}
	}

	c.Status(http.StatusNoContent)
}

// handlePostResourceError handles common post-related errors
func (h *V1Handler) handlePostResourceError(
	c *gin.Context,
	err error,
	resourceErrorCode v1errors.ErrorCode,
) bool {
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodePostNotFound))
			return true
		default:
			_ = c.Error(v1errors.InternalServerError(err).WithCode(resourceErrorCode))
			return true
		}
	}
	return false
}

// Post Helper functions

// Conversion functions from service types to API types

func (h *V1Handler) convertPostSummaryToAPI(post service.PostSummaryDetail) Post {
	return Post{
		Id:      post.ID.String(),
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
		Id:        post.ID.String(),
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
		UserId:    user.UserID.String(),
		DisplayId: user.DisplayID,
		Nickname:  user.Nickname,
		Bio:       user.Bio,
	}
}
