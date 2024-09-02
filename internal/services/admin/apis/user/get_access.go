package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type GetAccessRep struct {
	Avatar       string   `json:"avatar"`
	Buttons      []string `json:"buttons"`
	Code         int      `json:"code"`
	DeptId       int      `json:"deptId"`
	Introduction string   `json:"introduction"`
	Name         string   `json:"name"`
	Permissions  []string `json:"permissions"`
	Roles        []string `json:"roles"`
	UserId       int      `json:"userId"`
	UserName     string   `json:"userName"`
}

func (api *UserAPI) GetAccess(c *gin.Context) {
	var rep GetAccessRep
	var err error
	defer func() {
		if err != nil {
			c.Error(err)
		}
	}()
	claims := core.ExtractClaims(c)

	err = api.makePermissions(claims, &rep)
	if err != nil {
		return
	}
	err = api.makeUser(claims, &rep)
	if err != nil {
		return
	}
	core.OKRep(rep).SendGin(c)
}

func (api *UserAPI) makePermissions(claims *types.JwtClaims, rep *GetAccessRep) error {

	rep.Roles = append(rep.Roles, claims.RoleKey)

	if claims.RoleName == "admin" || claims.RoleName == "系统管理员" {
		rep.Permissions = append(rep.Permissions, "*:*:*")
		rep.Buttons = append(rep.Buttons, "*:*:*")
		return nil
	}

	list, err := api.getRolePermissions(claims.RoleId)
	if err != nil {
		return err
	}
	rep.Permissions = list
	rep.Buttons = list
	return nil
}

func (api *UserAPI) makeUser(claims *types.JwtClaims, rep *GetAccessRep) error {

	user, err := api.userRepo.FindOne(repository.WithUserId(claims.UserId))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return core.NewApiBizErr(err).
				SetBizCode(global.BizUnauthorizedErr).
				SetMsg("用户信息已过期")
		}
		return errors.WithStack(err)
	}
	rep.UserName = user.NickName
	rep.UserId = user.UserId
	rep.DeptId = user.DeptId
	rep.Name = user.NickName
	return nil
}

func (api *UserAPI) getRolePermissions(roleId int) ([]string, error) {
	role, err := api.roleRepo.FindOne(
		repository.WithRoleId(roleId),
		repository.WithPreloadSysMenu(),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil
		}
		return nil, errors.WithStack(err)
	}

	permissions := make([]string, 0)
	for _, menu := range role.SysMenu {
		if menu.Permission != "" {
			permissions = append(permissions, menu.Permission)
		}
	}
	return permissions, nil
}
