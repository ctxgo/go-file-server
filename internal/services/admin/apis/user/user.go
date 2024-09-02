package user

import "go-file-server/internal/common/repository"

type UserAPI struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	menuRepo *repository.MenuRepository
}

func NewUserAPI(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	menuRepo *repository.MenuRepository,
) *UserAPI {
	return &UserAPI{
		userRepo: userRepo,
		roleRepo: roleRepo,
		menuRepo: menuRepo,
	}
}
