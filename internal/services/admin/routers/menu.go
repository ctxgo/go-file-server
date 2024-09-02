package routers

import (
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/menu"
)

func RegisterMenuRoutes(svc *types.SvcCtx, menuApi *menu.MenuApi) {
	api := svc.Router.Group("/menu")
	{
		api.GET("", menuApi.GetPage)
		api.GET("/role-tree/:roleId", menuApi.GetRoleMenuTree)

	}
}
