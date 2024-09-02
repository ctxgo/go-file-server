package middlewares

import (
	"go-file-server/internal/common/core"

	"github.com/gin-gonic/gin"
)

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors[0].Err
		core.HandlingErr(c, err)
	}
}
