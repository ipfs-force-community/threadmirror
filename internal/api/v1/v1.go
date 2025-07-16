package v1

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/internal/service"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n"
	"github.com/ipfs-force-community/threadmirror/pkg/jobq"
)

// Check interface implementation
var _ ServerInterface = (*V1Handler)(nil)

type V1Handler struct {
	logger             *slog.Logger
	mentionService     *service.MentionService
	threadService      *service.ThreadService
	translationService *service.TranslationService
	commonConfig       *config.CommonConfig
	serverConfig       *config.ServerConfig
	jobQueueClient     jobq.JobQueueClient
}

func NewV1Handler(
	mentionService *service.MentionService,
	threadService *service.ThreadService,
	translationService *service.TranslationService,
	logger *slog.Logger,
	commonConfig *config.CommonConfig,
	serverConfig *config.ServerConfig,
	jobQueueClient jobq.JobQueueClient,
) *V1Handler {
	return &V1Handler{
		mentionService:     mentionService,
		threadService:      threadService,
		translationService: translationService,
		commonConfig:       commonConfig,
		serverConfig:       serverConfig,
		jobQueueClient:     jobQueueClient,
		logger:             logger.With("api", "v1"),
	}
}

func T(c *gin.Context, messageID string, templateData ...any) string {
	return i18n.T(c, messageID, templateData...)
}
