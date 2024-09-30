package user

import (
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/cache"
)

type UserAPI struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	menuRepo *repository.MenuRepository
	cache    cache.AdapterCache
}

func NewUserAPI(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	menuRepo *repository.MenuRepository,
	cache cache.AdapterCache,
) *UserAPI {
	return &UserAPI{
		userRepo: userRepo,
		roleRepo: roleRepo,
		menuRepo: menuRepo,
		cache:    cache,
	}
}
