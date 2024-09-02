package dept

import (
	"go-file-server/internal/services/admin/models"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GetInfoRep models.SysDept

type GetInfoReq struct {
	Id int `uri:"id"`
}

func (api *DeptApi) GetInfo(c *gin.Context) {

	var getInfoReq GetInfoReq
	err := c.ShouldBindUri(&getInfoReq)
	if err != nil {
		c.Error(errors.WithStack(err))
		return
	}

	data, err := api.getInfo(getInfoReq)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(data).SendGin(c)
}

func (api *DeptApi) getInfo(getInfoReq GetInfoReq) (GetInfoRep, error) {
	var getInfoRep GetInfoRep
	data, err := api.deptRepo.FindOne(repository.WithByDeptId(getInfoReq.Id))
	if err != nil {
		return getInfoRep, err
	}
	return GetInfoRep(data), nil

}
