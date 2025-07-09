package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
)

// GetQrcode implements the /qrcode endpoint.
func (h *V1Handler) GetQrcode(c *gin.Context, params GetQrcodeParams) {
	if params.GetThreadId() == "" {
		HandleBadRequestError(c, fmt.Errorf("thread_id parameter is required"))
		return
	}
	img, err := comm.GenQrcode(h.commonConfig.ThreadURLTemplate, params.GetThreadId())
	if err != nil {
		HandleInternalServerError(c, err)
		return
	}
	c.Data(http.StatusOK, "image/png", img)
}
