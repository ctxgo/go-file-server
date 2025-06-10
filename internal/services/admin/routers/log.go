package routers

import (
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/log/login"
	"go-file-server/internal/services/admin/apis/log/opera"
)

func RegisterLogRoutes(svc *types.SvcCtx, loginApi *login.LoginAPI, operaApi *opera.OperaAPI) {
	logApiGroup := svc.Router.Group("/log")
	logApiGroup.Use(middlewares.AuthCheckRole(svc))

	{
		loginLogApiGroup := logApiGroup.Group("/login")
		loginLogApiGroup.DELETE("", loginApi.Delete)
		loginLogApiGroup.GET("", loginApi.GetPage)
	}

	{
		operaLogApiGroup := logApiGroup.Group("/opera")
		operaLogApiGroup.DELETE("", operaApi.Delete)
		operaLogApiGroup.GET("", operaApi.GetPage)
	}

}
