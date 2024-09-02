package app

import (
	Init "go-file-server/internal/app/init"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin"
	"go-file-server/internal/services/normal"

	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"

	"github.com/gin-gonic/gin"
)

func Start() {
	//初始化组件
	svcCtx := Init.Initializer()

	//init gin
	ginEngine := Init.InitGin()
	svcCtx.Router = ginEngine

	//set gin
	setPublicMiddlewares(svcCtx)
	registerRouter(svcCtx)

	run(ginEngine)
}

// 公共中间件注册
func setPublicMiddlewares(svcCtx *types.SvcCtx) {
	r := svcCtx.Router
	r.Use(
		middlewares.NewGlog(repository.NewOperaLogRepository(svcCtx.Db)).GinLogger(),
	)
	r.Use(middlewares.ErrorHandlingMiddleware())

	r.Use(middlewares.GinRecovery())

	r.Use(middlewares.CORSMiddleware())
}

func registerRouter(svcCtx *types.SvcCtx) {
	admin.RegisterRouter(svcCtx)
	normal.RegisterRouter(svcCtx)
}

func run(ginEngine *gin.Engine) {
	listenport := config.ApplicationCfg.Port
	zlog.SugLog.Infof("******服务初始化完成,监听端口为: %v******", listenport)
	err := ginEngine.Run("0.0.0.0:" + listenport)
	if err != nil {
		zlog.SugLog.Fatal(err)
	}
}
