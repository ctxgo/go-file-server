package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UpdateDeptRep struct {
	UserId int `json:"userId" comment:"用户ID"`
}

type UpdateDeptReq struct {
	UserId int `json:"userId" comment:"用户ID"`
	CreateReq
}

func (api *UserAPI) Update(c *gin.Context) {
	var updateDeptReq UpdateDeptReq
	err := c.ShouldBind(&updateDeptReq)
	if err != nil {
		c.Error(err)
		return
	}
	err = api.updateUser(c, updateDeptReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)

}
func (api *UserAPI) updateUser(c *gin.Context, updateDeptReq UpdateDeptReq) error {
	userInfo, err := api.userRepo.FindOne(repository.WithUserId(updateDeptReq.UserId))
	if err != nil {
		return err
	}

	if userInfo.RoleId != updateDeptReq.RoleId {
		err = middlewares.UpdateLastTokenReset(api.cache, userInfo.UserId)
		if err != nil {
			return err
		}
	}

	claims := core.ExtractClaims(c)
	err = api.userRepo.Update(func(su *models.SysUser) {
		su.Username = updateDeptReq.Username
		su.Password = updateDeptReq.Password
		su.NickName = updateDeptReq.NickName
		su.Phone = updateDeptReq.Phone
		su.RoleId = updateDeptReq.RoleId
		su.Sex = updateDeptReq.Sex
		su.DeptId = updateDeptReq.DeptId
		su.Email = updateDeptReq.Email
		su.Status = updateDeptReq.Status
		su.Remark = updateDeptReq.Remark
		su.UpdateBy = claims.UserId
	}, repository.WithUserId(updateDeptReq.UserId))

	return errors.WithStack(err)

}
