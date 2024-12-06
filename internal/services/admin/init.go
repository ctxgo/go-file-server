package admin

import (
	"context"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/routers"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/pathtool"
	"go-file-server/pkgs/zlog"

	"github.com/casbin/casbin/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"gorm.io/gorm"
)

func setupSvcCtx(svcCtx *types.SvcCtx) *types.SvcCtx {
	gr := svcCtx.Router.Group("/api/v1")
	newCtx := svcCtx.Clone()
	newCtx.Router = gr
	return newCtx

}

func SetupSvcCtx(svcCtx *types.SvcCtx) *types.SvcCtx {
	ctx := setupSvcCtx(svcCtx)
	authenticator := middlewares.NewAuthenticator(
		repository.NewUserTokenRepository(ctx.Db),
		ctx.Cache,
	)
	ctx.Router.Use(middlewares.Auth(authenticator))
	return ctx
}

// 用于RegisterFsRoutes，该路由在内部鉴权
func SetupFsSvcCtx(svcCtx *types.SvcCtx) routers.FsSvcCtx {
	return setupSvcCtx(svcCtx)
}

func RegisterRouter(svcCtx *types.SvcCtx) {
	app := fx.New(
		fx.Provide(
			func() *gorm.DB { return svcCtx.Db },
		),
		fx.Provide(
			func() *pathtool.FileIndexer { return svcCtx.FsIndexer },
		),
		fx.Provide(
			func() *casbin.CachedEnforcer { return svcCtx.CasbinEnforcer },
		),
		fx.Provide(
			func() cache.AdapterCache { return svcCtx.Cache },
		),
		fx.Provide(
			func() *types.SvcCtx { return SetupSvcCtx(svcCtx) },
		),
		fx.Provide(
			func() routers.FsSvcCtx { return SetupFsSvcCtx(svcCtx) },
		),
		//路由
		routers.Routers,
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
	)
	if err := app.Start(context.Background()); err != nil {
		zlog.SugLog.Fatal(err)
	}
	if err := app.Stop(context.Background()); err != nil {
		zlog.SugLog.Fatal(err)
	}
}
