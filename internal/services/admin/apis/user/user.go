package user

import (
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/cache"
)

type UserAPI struct {
	userRepo      *repository.UserRepository
	userTokenRepo *repository.UserTokenRepository
	roleRepo      *repository.RoleRepository
	menuRepo      *repository.MenuRepository
	cache         cache.AdapterCache
}

func NewUserAPI(
	userRepo *repository.UserRepository,
	userTokenRepo *repository.UserTokenRepository,
	roleRepo *repository.RoleRepository,
	menuRepo *repository.MenuRepository,
	cache cache.AdapterCache,
) *UserAPI {
	return &UserAPI{
		userRepo:      userRepo,
		userTokenRepo: userTokenRepo,
		roleRepo:      roleRepo,
		menuRepo:      menuRepo,
		cache:         cache,
	}
}
