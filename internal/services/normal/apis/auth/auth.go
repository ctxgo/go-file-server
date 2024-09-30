package auth

import (
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/cache"
)

type Authenticator struct {
	deptRepo     *repository.DeptRepository
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	loginLogRepo *repository.LoginLogRepository
	cache        cache.AdapterCache
}

func NewAuthenticatorApi(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	loginLogRepo *repository.LoginLogRepository,
	deptRepo *repository.DeptRepository,
	cache cache.AdapterCache,
) *Authenticator {
	return &Authenticator{
		cache:        cache,
		deptRepo:     deptRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		loginLogRepo: loginLogRepo,
	}
}
