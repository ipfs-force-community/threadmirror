package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1errors "github.com/ipfs-force-community/threadmirror/internal/api/v1/errors"
	"gorm.io/datatypes"
)

type PaginationParm interface {
	GetLimit() *PageLimit
	GetOffset() *PageOffset
}

func ExtractPaginationParams(p PaginationParm) (limit int, offset int) {
	limit, offset = 20, 0
	if p.GetLimit() != nil {
		limit = *p.GetLimit()
	}

	if p.GetOffset() != nil {
		offset = *p.GetOffset()
	}

	return
}

func PaginatedJSON(c *gin.Context, data any, total int64, limit, offset int) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": struct {
			Total  int64 `json:"total"`
			Limit  int   `json:"limit"`
			Offset int   `json:"offset"`
		}{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func ParseUserID(c *gin.Context, id string) (datatypes.UUID, bool) {
	return ParseStringUUID(c, id, ErrCodeUserNotFound)
}

func ParseStringUUID(c *gin.Context, id string, errCode v1errors.ErrorCode) (datatypes.UUID, bool) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		_ = c.Error(v1errors.NotFound(err).WithCode(errCode))
		return datatypes.UUID{}, false
	}
	return datatypes.UUID(uuid), true
}
