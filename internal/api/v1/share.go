package v1

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
)

// GetShare implements the /share endpoint.
func (h *V1Handler) GetShare(c *gin.Context, params GetShareParams) {
	threadID := params.GetThreadId()
	if threadID == "" {
		_ = c.Error(v1errors.BadRequest(nil).WithCode(v1errors.ErrCodeBadRequest))
		return
	}

	// Fetch thread detail
	thread, err := h.threadService.GetThreadByID(c.Request.Context(), threadID)
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(v1errors.ErrCodeInternalError))
		return
	}
	if thread == nil {
		_ = c.Error(v1errors.NotFound(nil).WithCode(v1errors.ErrCodeNotFound))
		return
	}

	// Render thread to HTML
	html, err := comm.RenderThread(h.commonConfig.ThreadURLTemplate, threadID, thread, h.logger)
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(v1errors.ErrCodeInternalError))
		return
	}

	// Take screenshot using chromedp
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.DisableGPU,
	)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.EmulateViewport(485, 0),
		chromedp.Navigate("data:text/html;base64,"+base64.StdEncoding.EncodeToString([]byte(html))),
		chromedp.Sleep(1*time.Second),
		chromedp.FullScreenshot(&buf, 100),
	); err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(v1errors.ErrCodeInternalError))
		return
	}

	c.Data(http.StatusOK, "image/png", buf)
}
