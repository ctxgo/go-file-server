package routers

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/dept"
)

func RegisterDeptRoutes(svc *types.SvcCtx, deptApi *dept.DeptApi) {
	api := svc.Router.Group("/dept").Use(middlewares.AuthCheckRole(svc))
	{
		api.POST("", deptApi.Create)
		api.DELETE("", deptApi.Delete)
		api.PUT("", deptApi.UpdateDept)
		api.GET("", core.PermissionAction(svc.Db), deptApi.GetPage)
		api.GET(":id", deptApi.GetInfo)
		api.GET("tree", core.PermissionAction(svc.Db), deptApi.GetTree)
		api.GET("/role-tree/:roleId", deptApi.GetRoleDeptTree)

	}
}
