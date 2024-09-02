package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GetInfoRep *models.SysUser

type GetInfoReq struct {
	Id int `uri:"id"`
}

func (api *UserAPI) GetInfo(c *gin.Context) {

	var getInfoReq GetInfoReq
	err := c.ShouldBindUri(&getInfoReq)
	if err != nil {
		c.Error(err)
		return
	}

	data, err := api.getInfo(getInfoReq)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(data).SendGin(c)
}

func (api *UserAPI) getInfo(getInfoReq GetInfoReq) (GetInfoRep, error) {
	data, err := api.userRepo.FindOne(repository.WithUserId(getInfoReq.Id))
	return data, errors.WithStack(err)
}
