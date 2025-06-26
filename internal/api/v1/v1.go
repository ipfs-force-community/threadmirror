package v1

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/internal/bot"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n"
)

var _ ServerInterface = (*V1Handler)(nil)

type V1Handler struct {
	logger      *slog.Logger
	postService *service.PostService
	twitterBot  *bot.TwitterBot
}

func NewV1Handler(
	postService *service.PostService,
	logger *slog.Logger,
	twitterBot *bot.TwitterBot,
) *V1Handler {
	return &V1Handler{
		postService: postService,
		twitterBot:  twitterBot,
		logger:      logger.With("api", "v1"),
	}
}

func T(c *gin.Context, messageID string, templateData ...any) string {
	return i18n.T(c, messageID, templateData...)
}
