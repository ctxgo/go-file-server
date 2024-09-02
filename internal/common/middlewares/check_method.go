package middlewares

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"

	"github.com/gin-gonic/gin"
)

func AssertMethod(method string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == method {
			return
		}
		core.NewApiBizErr(nil).
			SetHttpCode(global.BadRequestError).
			SetBizCode(global.BizBadRequest).
			SetMsg("please use " + method + "...").
			SendGin(c)
	}
}
