package normal

import (
	"context"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/normal/routers"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/zlog"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"gorm.io/gorm"
)

func SetupGin(c *types.SvcCtx) gin.IRouter {
	gr := c.Router.Group("/api/v1")
	return gr
}

func RegisterRouter(svcCtx *types.SvcCtx) {
	app := fx.New(
		fx.Provide(
			func() *gorm.DB { return svcCtx.Db },
		),
		fx.Provide(
			func() gin.IRouter { return SetupGin(svcCtx) },
		),
		fx.Provide(
			func() cache.AdapterCache { return svcCtx.Cache },
		),
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
