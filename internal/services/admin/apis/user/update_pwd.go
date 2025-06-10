package user

import (
	"fmt"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type UpdatePwdReq struct {
	UserId      int    `json:"userId" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=1"`
	OldPassword string `json:"oldPassword" binding:"required,min=1"`
}

func (api *UserAPI) UpdatePwd(c *gin.Context) {
	var req UpdatePwdReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.Error(err)
		return
	}

	err = core.VerifyResourceOwner(c, req.UserId)
	if err != nil {
		c.Error(err)
		return
	}

	err = api.updatePwd(req)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)
}

func (api *UserAPI) updatePwd(req UpdatePwdReq) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		req.NewPassword = string(hash)
	}

	user, err := api.userRepo.FindOne(repository.WithUserId(req.UserId))
	if err != nil {
		return errors.WithStack(err)
	}
	ok, err := core.CompareHashAndPassword(user.Password, req.OldPassword)
	if err != nil || !ok {
		err = fmt.Errorf("CompareHashAndPassword error, %w", err)
		return core.NewApiBizErr(err).SetMsg("密码错误")
	}
	if user.Password == req.NewPassword {
		return nil
	}

	err = middlewares.UpdateLastTokenReset(api.cache, user.UserId)
	if err != nil {
		return err
	}

	err = api.userRepo.Update(func(su *models.SysUser) {
		su.Password = req.NewPassword
	}, repository.WithUserId(req.UserId))
	return errors.WithStack(err)

}
