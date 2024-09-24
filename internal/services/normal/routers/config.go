package routers

import (
	"go-file-server/internal/services/normal/apis/config"

	"github.com/gin-gonic/gin"
)

func RegisterConfigRoutes(r gin.IRouter, h config.ConfigHandler) {
	{
		r.GET("/config", gin.HandlerFunc(h))
	}
}
