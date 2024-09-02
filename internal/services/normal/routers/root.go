package routers

import (
	"go-file-server/internal/services/normal/apis/root"

	"github.com/gin-gonic/gin"
)

func RegisterRootRoutes(r gin.IRouter, h root.RootHandler) {
	{
		r.Any("", gin.HandlerFunc(h))
	}
}
