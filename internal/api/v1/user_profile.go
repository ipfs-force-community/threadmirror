package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var (
	// User profile module error codes: 14000-14999
	ErrCodeUserProfile = v1errors.NewErrorCode(v1errors.CheckCode(14000), "User profile error")

	// User profile operation errors
	ErrCodeFailedToGetUserProfile = v1errors.NewErrorCode(14001, "failed to get user profile")

	// User profile access errors
	ErrCodeUserNotFound = v1errors.NewErrorCode(14005, "user not found")
)

// User profile-related methods for V1Handler

// GetMe implements GET /me
func (h *V1Handler) GetMe(c *gin.Context) {
	authInfo := auth.MustAuthInfo(c)

	profile, err := h.userService.GetUserProfile(authInfo.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			_ = c.Error(v1errors.NotFound(err).WithCode(ErrCodeUserNotFound))
			return
		}
		_ = c.Error(v1errors.InternalServerError(err).WithCode(ErrCodeFailedToGetUserProfile))
		return
	}

	// Convert service types to API types
	apiProfile := h.serviceProfileToAPI(profile)

	c.JSON(http.StatusOK, gin.H{
		"data": apiProfile,
	})
}

// User Helper functions

func (h *V1Handler) serviceProfileToAPI(
	profile *service.UserProfileDetail,
) UserProfile {
	return UserProfile{
		Id:         profile.ID,
		DisplayId:  profile.DisplayID,
		Nickname:   profile.Nickname,
		Bio:        profile.Bio,
		Email:      (*openapi_types.Email)(profile.Email),
		PostsCount: profile.PostsCount,
		CreatedAt:  profile.CreatedAt,
		UpdatedAt:  profile.UpdatedAt,
	}
}
