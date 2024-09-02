package root

import (
	"go-file-server/internal/common/core"

	"github.com/gin-gonic/gin"
)

type RootHandler gin.HandlerFunc

func NewRootHandler() RootHandler {
	return func(c *gin.Context) {
		core.OKRep(nil).
			SendGin(c)
	}
}
