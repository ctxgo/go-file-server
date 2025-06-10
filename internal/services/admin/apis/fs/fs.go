package fs

import (
	"fmt"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/internal/services/admin/apis/role"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/pathtool"
	"go-file-server/pkgs/utils/limiter"
	resourcemanager "go-file-server/pkgs/utils/resource-manager"
	"go-file-server/pkgs/zlog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/pkg/errors"
)

type FsApi struct {
	Authenticator  *middlewares.Authenticator
	roleRepo       *repository.RoleRepository
	fsRepo         *repository.FsRepository
	casbinEnforcer *casbin.CachedEnforcer
	cache          cache.AdapterCache
	//流量限速器，用于download.go下载文件限速
	limiterManager utils.LimiterManager
	//双向map, 用于获取下载链接时，缓存下载元数据和路径id的对应关系
	idManager utils.IdManager
	//消息发布器，用于unarchiver.go解压文件时向多个客户端推送解压日志
	publisherManager *resourcemanager.ResourceManager[*utils.Publisher[utils.Message]]
	sync.RWMutex
}

func NewFsApi(
	roleRepo *repository.RoleRepository,
	fsRepo *repository.FsRepository,
	casbinEnforcer *casbin.CachedEnforcer,
	cache cache.AdapterCache,

) *FsApi {
	return &FsApi{
		roleRepo:         roleRepo,
		fsRepo:           fsRepo,
		casbinEnforcer:   casbinEnforcer,
		cache:            cache,
		limiterManager:   *utils.NewLimiterManager(30*time.Minute, 30*time.Minute),
		publisherManager: resourcemanager.NewResourceManager[*utils.Publisher[utils.Message]](),
		idManager:        *utils.NewIdManager(3*time.Hour, 3*time.Hour),
	}
}
func (api *FsApi) execRename(realPath, destination string) error {
	if realPath == destination {
		return nil
	}
	err := api.fsRepo.Rename(realPath, destination)
	if err != nil {
		ok, err := utils.ParsePathErr(err)
		if ok {
			return core.NewApiBizErr(err).
				SetMsg(err.Error())
		}

		zlog.SugLog.Error(err)
		return errors.Errorf("内部错误")
	}
	return nil
}

func (api *FsApi) getLimiter(uid int, roleKey string) (*limiter.Limiter, error) {

	raleLimiteBytes, err := GetRaleLimiteBytes(roleKey, api.cache, api.roleRepo)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%d-%s", uid, roleKey)
	raleLimiter := api.limiterManager.GetLimiter(key, raleLimiteBytes)
	return raleLimiter, nil

}

func (api *FsApi) ensureTempDir(roleKey string) (string, error) {

	tempPath, err := EnsureTempDir(roleKey)
	if err != nil {
		return "", err
	}
	return tempPath, api.fsRepo.MkdirAll(tempPath, os.ModePerm)
}

func GetRaleLimiteBytes(roleKey string, ch cache.AdapterCache,
	roleRepo *repository.RoleRepository) (uint64, error) {

	key := fmt.Sprintf("%s-%s", role.RateLimitKey, roleKey)
	rateLimitStr, err := ch.Get(key)
	if err == nil {
		return strconv.ParseUint(rateLimitStr, 10, 64)
	}
	if !cache.IsKeyNotFoundError(err) {
		return 0, err
	}
	data, err := roleRepo.FindOne(repository.WithRoleKey(roleKey))
	if err != nil {
		return 0, err
	}
	err = ch.Set(
		fmt.Sprintf("%s-%s", roleKey, role.RateLimitKey),
		data.RateLimit,
		0,
	)
	return data.RateLimit, err
}

func EnsureTempDir(roleKey string) (string, error) {
	tempPath, err := utils.GetRealPath(".tmp", roleKey)
	if err != nil {
		return "", core.NewApiBizErr(err).SetMsg(err.Error())
	}

	isExist, err := pathtool.NewFiletool(tempPath).IsExist()
	if err != nil {
		ok, err := utils.ParsePathErr(err)
		if ok {
			return "", core.NewApiBizErr(err).SetMsg(err.Error())
		}
		return "", err
	}
	if isExist {
		return tempPath, nil
	}
	return tempPath, nil
}
