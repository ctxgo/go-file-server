package opera

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
	Items []models.SysOperaLog `json:"items"`
}

type GetPageReq struct {
	types.Pagination
	Status    string `form:"status" `
	BeginTime string `form:"beginTime" `
	EndTime   string `form:"endTime" `
}

func (api *OperaAPI) GetPage(c *gin.Context) {

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

func (api *OperaAPI) getPage(getPageReq GetPageReq) (GetPageRep, error) {
	var count int64

	querys := api.makeQuery(getPageReq)

	dbScopes := append(querys,
		repository.WithOperaPaginateById(getPageReq.PageIndex, getPageReq.PageSize),
	)
	data, count, err := api.operarepo.Find(
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

func (api *OperaAPI) makeQuery(getPageReq GetPageReq) []base.DbScope {
	var querys []base.DbScope
	if getPageReq.Status != "" {
		querys = append(querys, repository.WithOperaStatus(getPageReq.Status))
	}
	if getPageReq.BeginTime != "" {
		querys = append(querys, repository.WithOperaBegin(getPageReq.BeginTime))
	}

	if getPageReq.EndTime != "" {
		querys = append(querys, repository.WithOperaEnd(getPageReq.EndTime))
	}

	return querys
}
