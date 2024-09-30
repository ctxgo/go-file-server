package role

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/base"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UpdateStatusRep struct {
	RoleId int `json:"roleId" comment:"角色编码"` // 角色编码
}

type UpdateStatusReq struct {
	RoleId int    `form:"roleId" binding:"required" comment:"角色编码"` // 角色编码
	Status string `form:"status" binding:"required" comment:"状态"`   // 状态
}

func (api *RoleApi) UpdateStatus(c *gin.Context) {
	var updateReq UpdateStatusReq
	err := c.ShouldBind(&updateReq)
	if err != nil {
		c.Error(err)
		return
	}
	err = api.updateStatus(updateReq)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(nil).SendGin(c)
}

func (api *RoleApi) updateStatus(updateReq UpdateStatusReq) error {
	role, err := api.roleRepo.FindOne(
		repository.WithRoleId(updateReq.RoleId),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	role.Status = updateReq.Status
	err = api.roleRepo.Save(role, base.WithFullAssociations())
	return errors.WithStack(err)
}
