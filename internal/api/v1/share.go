package v1

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
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

	// Take screenshot using chromedp
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("hide-scrollbars", true),
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
		HandleInternalServerError(c, err)
		return
	}

	c.Data(http.StatusOK, "image/png", buf)
}
