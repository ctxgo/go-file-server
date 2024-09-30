package role

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"github.com/gin-gonic/gin"
)

type GetPageRep struct {
	types.Page
	Items []models.SysRole `json:"items"`
}

type GetPageReq struct {
	types.Pagination
	RoleId   *int   `form:"roleId"`
	RoleName string `form:"roleName"`
	RoleKey  string `form:"roleKey"`
	Status   string `form:"status"`
}

func (api *RoleApi) GetPage(c *gin.Context) {

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

func (api *RoleApi) getPage(getPageReq GetPageReq) (GetPageRep, error) {

	querys := api.makeQuery(getPageReq)

	dbScopes := append(querys,
		repository.WithPaginateByRoleId(getPageReq.PageIndex, getPageReq.PageSize),
	)
	roles, count, err := api.roleRepo.FindWithCount(
		dbScopes...,
	)
	if err != nil {
		return GetPageRep{}, err
	}
	return GetPageRep{
		Page:  types.NewPage(count, getPageReq.PageIndex, getPageReq.PageSize),
		Items: roles,
	}, err

}

func (api *RoleApi) makeQuery(getPageReq GetPageReq) []base.DbScope {
	var querys []base.DbScope
	if getPageReq.RoleId != nil {
		querys = append(querys, repository.WithRoleId(*getPageReq.RoleId))
	}
	if getPageReq.RoleName != "" {
		querys = append(querys, repository.WithRoleName(getPageReq.RoleName))
	}
	if getPageReq.Status != "" {
		querys = append(querys, repository.WithUserStatus(getPageReq.Status))
	}
	if getPageReq.RoleKey != "" {
		querys = append(querys, repository.WithUsername(getPageReq.RoleKey))
	}

	return querys
}
