package routers

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/user"
)

func RegisterUserRoutes(svc *types.SvcCtx, userAPI *user.UserAPI) {
	api := svc.Router.Group("/user")
	{

		api.PUT("pwd", userAPI.UpdatePwd)
		api.POST("token", userAPI.GenToken)
		api.GET("token", userAPI.GetToken)
		api.DELETE("token", userAPI.DeleteToken)
		api.GET("access", userAPI.GetAccess)
		api.GET("profile", userAPI.GetProfile)
		api.GET("menu", userAPI.GetMenu)
	}

	authApi := svc.Router.Group("/user").Use(middlewares.AuthCheckRole(svc))
	{
		authApi.POST("", userAPI.Create)
		authApi.DELETE("", userAPI.Delete)
		authApi.PUT("", userAPI.Update)
		authApi.GET("", core.PermissionAction(svc.Db), userAPI.GetPage)
		authApi.GET(":id", userAPI.GetInfo)
	}
}
