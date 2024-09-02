package menu

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GetPageRep struct {
	Items models.SysMenuSlice `json:"items"`
}

type GetPageReq struct {
	Title   string `form:"title"  comment:"菜单名称"`   // 菜单名称
	Visible *int   `form:"visible"  comment:"显示状态"` // 显示状态
}

func (api *MenuApi) GetPage(c *gin.Context) {

	var getPageReq GetPageReq
	err := c.ShouldBind(&getPageReq)
	if err != nil {
		c.Error(err)
		return
	}
	data, err := api.getPage(getPageReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(data).SendGin(c)
}

func (api *MenuApi) getPage(getPageReq GetPageReq) (GetPageRep, error) {

	querys := api.makeQuery(getPageReq)

	dbScopes := append(querys,
		repository.WithOrderByMenuSort(false),
		repository.WithPreloadSysApi(),
	)
	menus, err := api.menuRepo.Find(
		dbScopes...,
	)
	if err != nil {
		return GetPageRep{}, errors.WithStack(err)
	}
	return GetPageRep{
		Items: api.buildMenuTree(menus, 0),
	}, err

}

func (api *MenuApi) makeQuery(getPageReq GetPageReq) []base.DbScope {
	var querys []base.DbScope
	if getPageReq.Title != "" {
		querys = append(querys, repository.WithMenuTitle(getPageReq.Title))
	}
	if getPageReq.Visible != nil {
		querys = append(querys, repository.WithMenuVisible(*getPageReq.Visible))
	}
	return querys
}

func (api *MenuApi) buildMenuTree(menus models.SysMenuSlice, parentId int) models.SysMenuSlice {
	var menusTree models.SysMenuSlice
	for _, menu := range menus {
		if menu.ParentId == parentId {
			menu.Children = api.buildMenuTree(menus, menu.MenuId)
			menusTree = append(menusTree, menu)
		}
	}
	return menusTree
}
