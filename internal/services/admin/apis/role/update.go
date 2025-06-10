package role

import (
	"go-file-server/internal/common/core"
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UpdateReq struct {
	RoleId    int    `uri:"id" binding:"required" comment:"角色编码"` // 角色编码
	DataScope string `form:"dataScope"`
	Admin     bool   `form:"admin" comment:"是否管理员"`
	MenuIds   []int  `form:"menuIds"`
	coreModels.ControlBy
	coreModels.ModelTime
	CreateReq
}

func (api *RoleApi) Update(c *gin.Context) {
	var updateDeptReq UpdateReq
	err := c.ShouldBind(&updateDeptReq)
	if err != nil {
		c.Error(err)
		return
	}
	err = api.updateRole(c, updateDeptReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)
}
func (api *RoleApi) updateRole(c *gin.Context, updateReq UpdateReq) error {
	claims := core.ExtractClaims(c)

	data := &models.SysRole{
		RoleId:    updateReq.RoleId,
		RoleKey:   updateReq.RoleKey,
		RoleName:  updateReq.RoleName,
		Status:    updateReq.Status,
		RoleSort:  updateReq.RoleSort,
		Remark:    updateReq.Remark,
		Admin:     updateReq.Admin,
		DataScope: updateReq.DataScope,

		ControlBy: coreModels.ControlBy{
			CreateBy: updateReq.CreateBy,
			UpdateBy: claims.UserId,
		},
		ModelTime: coreModels.ModelTime{
			CreatedAt: updateReq.CreatedAt,
			UpdatedAt: time.Now(),
			DeletedAt: updateReq.DeletedAt,
		},
	}

	err := api.cascadeUpdate(updateReq, data)
	if err != nil {
		return err
	}
	return api.updatePolicies(data.SysMenu, data.RoleKey)

}

func (api *RoleApi) cascadeUpdate(updateReq UpdateReq, data *models.SysRole) error {

	role, err := api.roleRepo.FindOne(
		repository.WithPreloadSysMenu(),
		repository.WithRoleId(updateReq.RoleId),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	menus, err := api.menuRepo.Find(
		repository.WithPreloadSysApi(),
		repository.WithMenuIds(updateReq.MenuIds...),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	err = api.roleRepo.Repo.WithTransaction(
		func(tx *gorm.DB) error {
			txRoleRepository := repository.NewRoleRepository(tx)
			err := txRoleRepository.
				DelWithAssociationSysMenu(role)
			if err != nil {
				return err
			}
			data.SysMenu = menus
			return txRoleRepository.Save(data, base.WithFullAssociations())
		},
	)

	return errors.WithStack(err)

}

func (api *RoleApi) updatePolicies(dataMenus models.SysMenuSlice, roleKey string) error {
	_, err := api.casbinEnforcer.RemoveFilteredPolicy(0, roleKey, "", "", models.AdminRoleKey)
	if err != nil {
		return err
	}
	return api.makePolicies(dataMenus, roleKey)
}
