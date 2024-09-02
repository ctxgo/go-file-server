package routers

import (
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/system"
)

func RegisterSystemRoutes(svc *types.SvcCtx, systemApi *system.SystemApi) {
	api := svc.Router.Group("/sse/")
	{
		api.GET("system", systemApi.GetInfo)
	}
}
