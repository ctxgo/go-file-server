package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type GetMenuRep models.SysMenuSlice

func (api *UserAPI) GetMenu(c *gin.Context) {

	sysMenus, err := api.makeMenu(c)

	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(sysMenus).SendGin(c)
}

func (api *UserAPI) makeMenu(c *gin.Context) (models.SysMenuSlice, error) {
	var sysMenus models.SysMenuSlice
	claims := core.ExtractClaims(c)

	sysMenuSlice, err := api.getMenu(claims.RoleKey, claims.RoleId)
	if err != nil {
		return sysMenus, errors.WithStack(err)
	}
	for _, menu := range sysMenuSlice {
		if menu.ParentId == 0 { // 只选择顶级菜单项开始构建
			menu.Children = buildMenuTree(sysMenuSlice, menu.MenuId)
			sysMenus = append(sysMenus, menu)
		}
	}
	return sysMenus, nil
}

func (api *UserAPI) getMenu(roleKey string, roleId int) (models.SysMenuSlice, error) {
	if roleKey == models.AdminRoleKey {
		return api.menuRepo.Find(
			repository.WithOrderByMenuSort(false),
			repository.WithMenuTypes(models.Directory, models.Menu),
			repository.WithDeletedAtIsNull(true))
	}
	role, err := api.roleRepo.FindOne(repository.WithRoleId(roleId), repository.WithPreloadSysMenu())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.SysMenuSlice{}, nil
		}
		return nil, errors.WithStack(err)
	}
	var roleMenus models.SysMenuSlice

	if role.SysMenu == nil {
		return roleMenus, nil
	}
	mIds := make([]int, 0)
	for _, menu := range role.SysMenu {
		mIds = append(mIds, menu.MenuId)
	}
	if err := api.recursiveSetMenu(mIds, &roleMenus); err != nil {
		return nil, err
	}

	roleMenus = roleMenus.RemoveDuplicatesByKey(func(sm *models.SysMenu) any { return sm.MenuId })
	return roleMenus, nil
}

func (api *UserAPI) recursiveSetMenu(mIds []int, menus *models.SysMenuSlice) error {
	if len(mIds) == 0 || menus == nil {
		return nil
	}
	roleMenus, err := api.menuRepo.Find(
		repository.WithMenuTypes(models.Directory, models.Menu, models.Button),
		repository.WithDeletedAtIsNull(true),
		repository.WithMenuIds(mIds...),
		repository.WithOrderByRoleSort(false),
	)

	if err != nil {
		return errors.WithStack(err)
	}

	subIds := make([]int, 0)
	for _, menu := range roleMenus {
		if menu.ParentId != 0 {
			subIds = append(subIds, menu.ParentId)
		}
		if menu.MenuType != models.Button {
			*menus = append(*menus, menu)
		}
	}
	return api.recursiveSetMenu(subIds, menus)
}

// buildMenuTree 通过递归构建每个菜单项的子树
func buildMenuTree(menus models.SysMenuSlice, parentId int) models.SysMenuSlice {
	var children models.SysMenuSlice
	for _, menu := range menus {
		if menu.ParentId != parentId {
			continue
		}
		if menu.MenuType != models.Button { // 只有非"Button"类型的菜单项才继续构建子树
			menu.Children = buildMenuTree(menus, menu.MenuId)
		}
		children = append(children, menu)
	}
	return children
}
