package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"github.com/ipfs-force-community/threadmirror/internal/comm"
)

// GetQrcode implements the /qrcode endpoint.
func (h *V1Handler) GetQrcode(c *gin.Context, params GetQrcodeParams) {
	if params.GetThreadId() == "" {
		_ = c.Error(v1errors.BadRequest(nil).WithCode(v1errors.ErrCodeBadRequest))
		return
	}
	img, err := comm.GenQrcode(h.commonConfig.ThreadURLTemplate, params.GetThreadId())
	if err != nil {
		_ = c.Error(v1errors.InternalServerError(err).WithCode(v1errors.ErrCodeInternalError))
		return
	}
	c.Data(http.StatusOK, "image/png", img)
}
