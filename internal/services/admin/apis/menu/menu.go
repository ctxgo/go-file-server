package menu

import "go-file-server/internal/common/repository"

type MenuApi struct {
	menuRepo *repository.MenuRepository
	roleRepo *repository.RoleRepository
}

func NewRoleApi(
	menuRepo *repository.MenuRepository,
	roleRepo *repository.RoleRepository,
) *MenuApi {
	return &MenuApi{
		menuRepo: menuRepo,
		roleRepo: roleRepo,
	}
}
