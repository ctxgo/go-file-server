package role

import (
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/cache"

	"github.com/casbin/casbin/v2"
)

type RoleApi struct {
	userRepo       *repository.UserRepository
	roleRepo       *repository.RoleRepository
	menuRepo       *repository.MenuRepository
	deptRepo       *repository.DeptRepository
	casbinEnforcer *casbin.CachedEnforcer
	cache          cache.AdapterCache
}

func NewRoleApi(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	menuRepo *repository.MenuRepository,
	deptRepo *repository.DeptRepository,
	casbinEnforcer *casbin.CachedEnforcer,
	cache cache.AdapterCache,
) *RoleApi {
	return &RoleApi{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		menuRepo:       menuRepo,
		deptRepo:       deptRepo,
		casbinEnforcer: casbinEnforcer,
		cache:          cache,
	}
}
