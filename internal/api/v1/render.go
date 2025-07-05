package v1

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
)

// GetRender implements the /render endpoint.
func (h *V1Handler) GetRender(c *gin.Context, params GetRenderParams) {
	if params.GetThreadId() == "" {
		_ = c.Error(v1errors.BadRequest(nil).WithCode(v1errors.ErrCodeBadRequest))
		return
	}

	thread, err := h.threadService.GetThreadByID(c.Request.Context(), params.GetThreadId())
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(v1errors.ErrCodeInternalError))
		return
	}
	if thread == nil {
		_ = c.Error(v1errors.NotFound(err).WithCode(v1errors.ErrCodeNotFound))
		return
	}

	c.Render(http.StatusOK, &render{
		threadURLTemplate: h.commonConfig.ThreadURLTemplate,
		threadID:          params.GetThreadId(),
		data:              thread,
		logger:            h.logger,
	})
}

type render struct {
	threadURLTemplate string
	threadID          string
	data              any
	logger            *slog.Logger
}

func (r *render) Render(w http.ResponseWriter) error {
	html, err := comm.RenderThread(r.threadURLTemplate, r.threadID, r.data, r.logger)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(html))
	return err
}

func (r *render) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
}
