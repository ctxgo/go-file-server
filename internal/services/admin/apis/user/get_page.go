package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GetPageRep struct {
	types.Page
	Items []models.SysUser `json:"items"`
}

type GetPageReq struct {
	types.Pagination
	DeptId   string `form:"deptId"` //这里会传入 /DeptId/格式，模糊匹配DeptPath
	UserId   int    `form:"userId"`
	Username string `form:"username"`
	NickName string `form:"nickName"`
	Phone    string `form:"phone"`
	Status   string `form:"status"`
}

func (api *UserAPI) GetPage(c *gin.Context) {

	var getPageReq GetPageReq
	err := c.ShouldBind(&getPageReq)
	if err != nil {
		c.Error(err)
		return
	}
	data, err := api.getPage(c, getPageReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(data).SendGin(c)
}

func (api *UserAPI) getPage(c *gin.Context, getPageReq GetPageReq) (GetPageRep, error) {

	querys := api.makeQuery(getPageReq)

	dbScopes := append(querys,
		repository.WithPreloadDept(),
		core.WithUserPermissionDbScope(repository.GetTableName(), c),
		repository.WithPaginateByUserId(getPageReq.PageIndex, getPageReq.PageSize),
	)
	users, count, err := api.userRepo.Find(
		dbScopes...,
	)
	if err != nil {
		return GetPageRep{}, errors.WithStack(err)
	}
	return GetPageRep{
		Page:  types.NewPage(count, getPageReq.PageIndex, getPageReq.PageSize),
		Items: users,
	}, err

}

func (api *UserAPI) makeQuery(getPageReq GetPageReq) []base.DbScope {
	var querys []base.DbScope
	if getPageReq.DeptId != "" {
		querys = append(querys,
			repository.WithJoinDeptOnDeptID(),
			repository.WithLikeDeptPath(getPageReq.DeptId),
		)
	}
	if getPageReq.Phone != "" {
		querys = append(querys, repository.WithPhone(getPageReq.Phone))
	}
	if getPageReq.Status != "" {
		querys = append(querys, repository.WithUserStatus(getPageReq.Status))
	}
	if getPageReq.Username != "" {
		querys = append(querys, repository.WithUsername(getPageReq.Username))
	}
	if getPageReq.Username != "" {
		querys = append(querys, repository.WithUsername(getPageReq.Username))
	}
	return querys
}
