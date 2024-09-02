package routers

import (
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/role"
)

func RegisterRoleRoutes(svc *types.SvcCtx, roleApi *role.RoleApi) {
	roleGrop := svc.Router.Group("/role").Use(middlewares.AuthCheckRole(svc))
	{

		roleGrop.PUT("datascope", roleApi.UpdateDataScope)
		roleGrop.PUT("status", roleApi.UpdateStatus)
		roleGrop.PUT("fs", roleApi.UpdateFs)
		roleGrop.GET(":id", roleApi.GetInfo)
		roleGrop.GET("", roleApi.GetPage)
		roleGrop.POST("", roleApi.Create)
		roleGrop.DELETE("", roleApi.Delete)
		roleGrop.PUT("", roleApi.Update)
	}

}
