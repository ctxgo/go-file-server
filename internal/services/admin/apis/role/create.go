package role

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/config"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type CreateRep struct {
	RoleId int `uri:"id" ` // 角色编码
}

type CreateReq struct {
	RoleName string           `form:"roleName" ` // 角色名称
	Status   string           `form:"status" `   // 状态 1禁用 2正常
	RoleKey  string           `form:"roleKey"`   // 角色代码
	RoleSort int              `form:"roleSort" ` // 角色排序
	Remark   string           `form:"remark" `   // 备注
	SysMenu  []models.SysMenu `form:"sysMenu"`
	MenuIds  []int            `form:"menuIds"`
}

func (api *RoleApi) Create(c *gin.Context) {

	var createVals CreateReq
	err := c.ShouldBind(&createVals)
	if err != nil {
		c.Error(err)
		return
	}

	roleId, err := api.create(createVals)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(CreateRep{
		RoleId: roleId,
	}).SendGin(c)

}

func (api *RoleApi) create(createReq CreateReq) (int, error) {

	dataMenus, err := api.menuRepo.Find(repository.WithPreloadSysApi(),
		repository.WithMenuIds(createReq.MenuIds...))
	if err != nil {
		return 0, errors.WithStack(err)
	}

	data := &models.SysRole{
		RoleName: createReq.RoleName,
		RoleKey:  createReq.RoleKey,
		Status:   createReq.Status,
		RoleSort: createReq.RoleSort,
		Remark:   createReq.Remark,
		MenuIds:  createReq.MenuIds,
		SysMenu:  dataMenus,
	}
	err = api.roleRepo.Repo.Create(data)
	if err != nil {
		if repository.IsDuplicateError(config.DatabaseCfg.Driver, err) {
			return 0, core.NewApiBizErr(errors.WithStack(err)).
				SetMsg("权限名称已经存在")

		}
		return 0, errors.WithStack(err)
	}
	err = api.makePolicies(dataMenus, data.RoleKey)
	return data.RoleId, err
}

func (api *RoleApi) makePolicies(dataMenus models.SysMenuSlice, roleKey string) error {
	mp := make(map[string]bool, 0)
	polices := make([][]string, 0)
	for _, menu := range dataMenus {
		for _, api := range menu.SysApi {
			if !mp[roleKey+"-"+api.Path+"-"+api.Action] {
				mp[roleKey+"-"+api.Path+"-"+api.Action] = true
				polices = append(polices,
					[]string{roleKey, api.Path, api.Action, models.AdminRoleKey})
			}
		}
	}
	if len(polices) <= 0 {
		return nil
	}
	_, err := api.casbinEnforcer.AddNamedPolicies("p", polices)
	if err != nil {
		return errors.WithStack(err)
	}
	err = api.casbinEnforcer.LoadPolicy()
	return errors.WithStack(err)
}
