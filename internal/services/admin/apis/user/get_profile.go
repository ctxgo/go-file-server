package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GetProfileRep struct {
	User  *models.SysUser  `json:"user"`
	Roles []models.SysRole `json:"roles"`
}

func (api *UserAPI) GetProfile(c *gin.Context) {

	claims := core.ExtractClaims(c)

	data, err := api.getProfile(claims.UserId)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(data).SendGin(c)
}

func (api *UserAPI) getProfile(userId int) (GetProfileRep, error) {
	var getProfileRep GetProfileRep
	user, err := api.userRepo.FindOne(
		repository.WithPreloadDept(),
		repository.WithUserId(userId))

	if err != nil {
		return getProfileRep, errors.WithStack(err)
	}
	getProfileRep.User = user
	roles, err := api.roleRepo.Find(repository.WithRoleId(user.RoleId))

	if err != nil {
		return getProfileRep, errors.WithStack(err)
	}
	getProfileRep.Roles = roles

	return getProfileRep, nil
}
