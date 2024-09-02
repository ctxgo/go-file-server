package menu

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type MenuLabel struct {
	Id       int         `json:"id,omitempty" gorm:"-"`
	Label    string      `json:"label,omitempty" gorm:"-"`
	Children []MenuLabel `json:"children,omitempty" gorm:"-"`
}

type GetTreeRep struct {
	Menus       []MenuLabel `json:"menus"`
	CheckedKeys []int       `json:"checkedKeys"`
}

type GetTreeReq struct {
	RoleId int `uri:"roleId"`
}

func (api *MenuApi) GetRoleMenuTree(c *gin.Context) {

	var getTreeReq GetTreeReq
	err := c.ShouldBind(&getTreeReq)
	if err != nil {
		c.Error(err)
		return
	}
	data, err := api.getTree(getTreeReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(data).SendGin(c)
}

func (api *MenuApi) getTree(getTreeReq GetTreeReq) (GetTreeRep, error) {
	var data GetTreeRep
	menus, err := api.menuRepo.Find(repository.WithOrderByMenuSort(false))
	if err != nil {
		return data, errors.WithStack(err)
	}
	data.Menus = api.buildMenuLableTree(menus, 0)
	if getTreeReq.RoleId != 0 {
		menuIds, err := api.getRoleMenuId(getTreeReq.RoleId)
		if err != nil {
			return data, errors.WithStack(err)
		}
		data.CheckedKeys = menuIds
	}
	return data, err

}

// GetRoleMenuId 获取角色对应的菜单ids
func (api *MenuApi) getRoleMenuId(roleId int) ([]int, error) {
	var data []int
	role, err := api.roleRepo.FindOne(repository.WithPreloadSysMenu(), repository.WithRoleId(roleId))
	if err != nil {
		return data, errors.WithStack(err)
	}

	l := role.SysMenu
	for i := 0; i < len(l); i++ {
		data = append(data, l[i].MenuId)
	}
	return data, nil
}

func (api *MenuApi) buildMenuLableTree(menus models.SysMenuSlice, parentId int) []MenuLabel {
	var menuLabelTree []MenuLabel
	for _, menu := range menus {
		if menu.ParentId == parentId {

			data := MenuLabel{
				Id:       menu.MenuId,
				Label:    menu.Title,
				Children: api.buildMenuLableTree(menus, menu.MenuId),
			}
			menuLabelTree = append(menuLabelTree, data)
		}
	}
	return menuLabelTree
}
