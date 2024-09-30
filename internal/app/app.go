package app

import (
	"context"
	"fmt"
	Init "go-file-server/internal/app/init"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin"
	"go-file-server/internal/services/normal"
	"net/http"
	"os"
	"syscall"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"

	"go-file-server/pkgs/config"
	"go-file-server/pkgs/zlog"

	"github.com/oklog/run"
)

func Start() error {
	//初始化组件
	svcCtx := Init.Initializer()
	var group = &run.Group{}

	if config.FptCfg.Enable {
		addFtpServer(group, svcCtx)
	}

	addGinServer(group, svcCtx)

	group.Add(run.SignalHandler(context.Background(), os.Interrupt, syscall.SIGTERM))
	if err := group.Run(); err != nil {
		if _, ok := err.(run.SignalError); !ok {
			return fmt.Errorf("run groups: %w", err)
		}
		zlog.SugLog.Error("shutdown now", "err", err)
	}
	return nil
}

func addFtpServer(group *run.Group, svcCtx *types.SvcCtx) {

	driver, err := Init.InitFtpServer(svcCtx)
	if err != nil {
		zlog.SugLog.Fatal(err)
	}
	ftpserver := ftpserver.NewFtpServer(driver)

	stop := func() {
		driver.Stop()
		if err := ftpserver.Stop(); err != nil {
			zlog.SugLog.Errorf(
				"Problem stopping ftp server, err: %v", err,
			)
		}
	}

	group.Add(
		func() error {
			zlog.SugLog.Infof(
				"******ftp服务初始化完成,监听地址为 %s******",
				config.FptCfg.Addr,
			)
			return ftpserver.ListenAndServe()
		},
		func(err error) {
			go stop()
			if err := driver.WaitGracefully(time.Second * 3); err != nil {
				zlog.SugLog.Errorf("stop ftp server err: %v", err)
			}
		},
	)
}

func addGinServer(group *run.Group, svcCtx *types.SvcCtx) {
	//init gin
	ginEngine := Init.InitGin()
	svcCtx.Router = ginEngine

	//set gin
	setPublicMiddlewares(svcCtx)
	registerRouter(svcCtx)

	host := fmt.Sprintf("%s:%s", config.ApplicationCfg.Host, config.ApplicationCfg.Port)
	srv := &http.Server{
		Addr:    host,
		Handler: ginEngine.Handler(),
	}
	group.Add(
		func() error {
			zlog.SugLog.Infof(
				"******gin服务初始化完成,监听地址为 %s******",
				host,
			)
			return srv.ListenAndServe()
		},
		func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				zlog.SugLog.Errorf("stop gin server err:%v", err)
			}
		},
	)
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
