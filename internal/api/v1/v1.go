package v1

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/internal/bot"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n"
)

var _ ServerInterface = (*V1Handler)(nil)

type V1Handler struct {
	logger         *slog.Logger
	supabaseConfig *config.SupabaseConfig
	userService    *service.UserService
	postService    *service.PostService
	twitterBot     *bot.TwitterBot
}

func NewV1Handler(
	userService *service.UserService,
	postService *service.PostService,
	supabaseConfig *config.SupabaseConfig,
	logger *slog.Logger,
	twitterBot *bot.TwitterBot,
) *V1Handler {
	return &V1Handler{
		supabaseConfig: supabaseConfig,
		userService:    userService,
		postService:    postService,
		twitterBot:     twitterBot,
		logger:         logger.With("api", "v1"),
	}
}

func T(c *gin.Context, messageID string, templateData ...any) string {
	return i18n.T(c, messageID, templateData...)
}
