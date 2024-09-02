package routers

import (
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/avatar"
)

func RegisterAvatarRoutes(svc *types.SvcCtx, avatarApi *avatar.AvatarAPI) {
	api := svc.Router.Group("/avatar")
	{
		api.POST("", avatarApi.Create)
		api.GET("", avatarApi.Get)

	}
}
