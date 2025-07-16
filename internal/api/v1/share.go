package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
	"github.com/ipfs-force-community/threadmirror/pkg/util"
)

// GetShare implements the /share endpoint.
func (h *V1Handler) GetShare(c *gin.Context, params GetShareParams) {
	threadID := params.GetThreadId()
	if threadID == "" {
		HandleBadRequestError(c, fmt.Errorf("thread_id parameter is required"))
		return
	}

	// Fetch thread detail
	thread, err := h.threadService.GetThreadByID(c.Request.Context(), threadID)
	if err != nil {
		HandleInternalServerError(c, err)
		return
	}
	if thread == nil {
		HandleNotFoundError(c, fmt.Errorf("thread not found"))
		return
	}

	// Render thread to HTML
	html, err := comm.RenderThread(h.commonConfig.ThreadURLTemplate, threadID, thread, h.logger)
	if err != nil {
		HandleInternalServerError(c, err)
		return
	}

	// Take screenshot using utility function
	buf, err := util.TakeScreenshotFromHTML(c.Request.Context(), string(html), nil)
	if err != nil {
		HandleInternalServerError(c, err)
		return
	}

	c.Data(http.StatusOK, "image/png", buf)
}
