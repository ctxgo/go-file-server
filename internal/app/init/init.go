package init

import (
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/cronjob"
	"go-file-server/internal/ftpserver"
	"go-file-server/internal/services/admin/apis/fs"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/casbin"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/pathtool"
	"go-file-server/pkgs/utils/captcha"
	"go-file-server/pkgs/utils/retry"
	"go-file-server/pkgs/utils/str"
	"go-file-server/pkgs/zlog"
	"os"

	"time"

	Casbin "github.com/casbin/casbin/v2"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func init() {
	str.InitSnowflake()
}

func Initializer() *types.SvcCtx {
	// 工作目录
	initBaseDir(config.ApplicationCfg.Basedir)

	// 数据库
	db, err := initDB()
	if err != nil {
		zlog.SugLog.Fatal(err)
	}

	// 初始化缓存
	cache, err := initCache()
	if err != nil {
		zlog.SugLog.Fatal(err)
	}

	//定时任务
	cronjob.InitJobs(db)

	// 初始化验证码组件
	initCaptcha(cache)

	// 权限组件
	casbinEnforcer, err := initCasbin(db, cache)
	if err != nil {
		zlog.SugLog.Fatal(err)
	}

	// bleve文档索引
	fsIndexer, err := initFileIndexer(cache)
	if err != nil {
		zlog.SugLog.Fatal(err)
	}
	return &types.SvcCtx{
		Db:             db,
		Cache:          cache,
		FsIndexer:      fsIndexer,
		CasbinEnforcer: casbinEnforcer,
	}
}

func initCache() (c cache.AdapterCache, err error) {
	if config.CacheCfg.Redis == nil {
		c = cache.NewMemory()
		return
	}
	retry.Retry(
		func() error {
			c, err = cache.NewRedis(&redis.Options{
				Addr:     config.CacheCfg.Redis.Addr,
				Password: config.CacheCfg.Redis.Password,
				DB:       config.CacheCfg.Redis.DB,
			})
			if err != nil {
				err = errors.Errorf("open redis faild , host: %s  err: %v",
					config.CacheCfg.Redis.Addr,
					err,
				)
			}
			return err
		},
	)
	return
}

func initCasbin(db *gorm.DB, cache cache.AdapterCache) (casbinEnforcer *Casbin.CachedEnforcer, err error) {
	ops := []casbin.Option{}
	ops = append(
		ops, casbin.WithGormDB(db),
		casbin.WithCasbinTablePrefix("sys"),
	)
	if cache.String() == "redis" {
		_client := cache.GetClient()
		client := _client.(*redis.Client)
		ops = append(ops, casbin.WithCache(casbin.NewRedisCache(client)))
	}

	casbinEnforcer, err = casbin.NewEnforcer(
		ops...,
	)
	if err != nil {
		zlog.SugLog.Fatal(err)
	}
	return
}

func initCaptcha(cache cache.AdapterCache) {
	captcha.SetStore(captcha.NewCacheStore(cache, 3*time.Minute))
}

func initFileIndexer(cache cache.AdapterCache) (*pathtool.FileIndexer, error) {
	updateCallback := func(f *pathtool.FileIndexer) {
		var data fs.InfoRep
		var err error
		defer func() {
			if err != nil {
				zlog.SugLog.Error(err)
				return
			}
			strData, err := str.ConvertToString(data)
			if err != nil {
				zlog.SugLog.Error(err)
			}
			err = cache.Set(fs.FsKey, strData, 0)
			if err != nil {
				zlog.SugLog.Error(err)
			}
		}()
		fsRepo := repository.NewFsRepository(f)
		data.Count, err = fsRepo.GetCount()
		if err != nil {
			return
		}
		data.DirCount, err = fsRepo.GetCount(repository.WithIsDir(true))
		if err != nil {
			return
		}
		data.FileCount = data.Count - data.DirCount
	}

	return pathtool.NewFileIndexer(
		config.ApplicationCfg.Basedir,
		pathtool.WithLog(zlog.SugLog),
		pathtool.WithStorageType(pathtool.UseDisk),
		pathtool.WithUpdateCallback(updateCallback),
	)
}

func initBaseDir(realPath string) {
	isdir, err := pathtool.NewFiletool(realPath).AssertDir()
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(realPath, os.ModePerm)
		}
		zlog.SugLog.Fatal(err)
	}

	if !isdir {
		zlog.SugLog.Fatal("工作目录不是标准的目录")
	}
}

func InitFtpServer(svcCtx *types.SvcCtx) (*ftpserver.Server, error) {
	ftpCfg := config.FptCfg

	return ftpserver.NewServer(
		svcCtx,
		ftpserver.WithAddr(ftpCfg.Addr),
		ftpserver.WithPublicHost(ftpCfg.PublicHost),
		ftpserver.WithPassivePortRange(ftpCfg.PassivePortStart, ftpCfg.PassivePortEnd),
		ftpserver.WithLogger(zlog.SugLog),
	)
}
