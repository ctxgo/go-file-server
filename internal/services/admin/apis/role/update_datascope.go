package role

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/pkgs/base"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UpdateDataScopeRep struct {
	RoleId int `json:"roleId" comment:"角色编码"` // 角色编码
}
type UpdateDataScopeReq struct {
	RoleId    int    `json:"roleId" binding:"required"`
	DataScope string `json:"dataScope" binding:"required"`
	DeptIds   []int  `json:"deptIds"`
}

func (api *RoleApi) UpdateDataScope(c *gin.Context) {
	var updateReq UpdateDataScopeReq
	err := c.ShouldBind(&updateReq)
	if err != nil {
		c.Error(err)
		return
	}
	err = api.updateDataScope(c, updateReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)
}
func (api *RoleApi) updateDataScope(c *gin.Context, updateReq UpdateDataScopeReq) error {
	claims := core.ExtractClaims(c)

	role, err := api.roleRepo.FindOne(repository.WithPreloadSysDept(),
		repository.WithRoleId(updateReq.RoleId),
	)
	if err != nil {
		return err
	}
	depts, err := api.deptRepo.Find(repository.WithDeptIds(updateReq.DeptIds...))
	if err != nil {
		return err
	}

	return api.roleRepo.Repo.WithTransaction(
		func(tx *gorm.DB) error {
			txRoleRepository := repository.NewRoleRepository(tx)
			err := txRoleRepository.
				DelWithAssociationSysDept(role)
			if err != nil {
				return errors.WithStack(err)
			}

			role.SysDept = depts
			role.DataScope = updateReq.DataScope
			role.UpdateBy = claims.UserId
			err = txRoleRepository.Save(role, base.WithFullAssociations())
			return errors.WithStack(err)

		},
	)
}
