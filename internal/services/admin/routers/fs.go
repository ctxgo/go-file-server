package routers

import (
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/fs"
)

type FsSvcCtx *types.SvcCtx

func RegisterFsRoutes(svc FsSvcCtx, fsApi *fs.FsApi) {

	authenticator := middlewares.NewAuthenticator(
		repository.NewUserTokenRepository(svc.Db),
		svc.Cache,
	)
	fsApi.Authenticator = authenticator

	router := svc.Router.Group("/fsd")
	{
		router.GET("/*path", fsApi.Download)
	}

	authRouter := svc.Router.Group("")
	authRouter.Use(middlewares.Auth(authenticator))

	{
		authRouter.GET("/sse/fs/info", fsApi.GetInfo)
		authRouter.GET("/sse/fs/unarchive/*path", fsApi.Unarchive)
		authRouter.PUT("/fs/*path", fsApi.Update)
		authRouter.GET("/fsu/*path", fsApi.GetDownloadUrl)
		authRouter.POST("/fsindex", fsApi.Reset)

	}

	fsCheckRoleGroup := authRouter.Group("/fs").Use(middlewares.AuthCheckRole(svc))
	{
		fsCheckRoleGroup.GET("*path", fsApi.GetPage)
		fsCheckRoleGroup.DELETE("*path", fsApi.Delete)
		fsCheckRoleGroup.POST("*path", fsApi.Create)
	}

}
