package login

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
	Items []models.SysLoginLog `json:"items"`
}

type GetPageReq struct {
	types.Pagination
	Username      string `form:"username" `
	Status        string `form:"status" `
	Ipaddr        string `form:"ipaddr" `
	LoginLocation string `form:"loginLocation" `
	BeginTime     string `form:"beginTime" `
	EndTime       string `form:"endTime" `
}

func (api *LoginAPI) GetPage(c *gin.Context) {

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

func (api *LoginAPI) getPage(getPageReq GetPageReq) (GetPageRep, error) {
	var count int64

	querys := api.makeQuery(getPageReq)

	dbScopes := append(querys,
		repository.WithLoginPaginateById(getPageReq.PageIndex, getPageReq.PageSize),
	)
	data, count, err := api.loginrepo.Find(
		dbScopes...,
	)
	if err != nil {
		return GetPageRep{}, err
	}

	return GetPageRep{
		Page:  types.NewPage(count, getPageReq.PageIndex, getPageReq.PageSize),
		Items: data,
	}, err

}

func (api *LoginAPI) makeQuery(getPageReq GetPageReq) []base.DbScope {
	var querys []base.DbScope
	if getPageReq.Username != "" {
		querys = append(querys, repository.WithLoginUsername(getPageReq.Username))
	}
	if getPageReq.Status != "" {
		querys = append(querys, repository.WithLoginStatus(getPageReq.Status))
	}
	if getPageReq.Ipaddr != "" {
		querys = append(querys, repository.WithLoginIpaddr(getPageReq.Ipaddr))
	}

	return querys
}
