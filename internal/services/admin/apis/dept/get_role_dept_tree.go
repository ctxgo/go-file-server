package dept

import (
	"go-file-server/internal/common/core"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type RoleDeptTreeRep struct {
	Depts       []*GetTreeData `json:"depts"`
	CheckedKeys []int          `json:"checkedKeys"`
}

type GetTreeReq struct {
	RoleId int `uri:"roleId"`
}

func (api *DeptApi) GetRoleDeptTree(c *gin.Context) {

	var getTreeReq GetTreeReq
	err := c.ShouldBindUri(&getTreeReq)
	if err != nil {
		c.Error(err)
		return
	}
	data, err := api.getRoleDeptTree(c, getTreeReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(data).SendGin(c)
}

func (api *DeptApi) getRoleDeptTree(c *gin.Context, getTreeReq GetTreeReq) (RoleDeptTreeRep, error) {
	var roleDeptTreeRep RoleDeptTreeRep
	roleDeptTree, err := api.makeTree(c)
	if err != nil {
		return roleDeptTreeRep, err
	}
	roleDeptTreeRep.Depts = roleDeptTree
	if getTreeReq.RoleId != 0 {
		deptIds, err := api.deptRepo.GetFilteredDeptIdsForRole(getTreeReq.RoleId)
		if err != nil {
			return roleDeptTreeRep, errors.WithStack(err)
		}
		roleDeptTreeRep.CheckedKeys = deptIds
	}

	return roleDeptTreeRep, err

}
