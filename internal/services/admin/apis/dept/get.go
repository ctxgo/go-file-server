package dept

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Data struct {
	*models.SysDept
	Children []*Data `json:"children,omitempty"`
}

type GetRep []*Data

func (api *DeptApi) GetPage(c *gin.Context) {
	depts, err := api.makeDeptPage(c)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(depts).SendGin(c)

}

func (api *DeptApi) makeDeptPage(c *gin.Context) (GetRep, error) {
	list, err := api.deptRepo.Find(core.WithDeptDbScope(c))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sysDept := api.buildDeptPage(list)
	return sysDept, nil
}

func (api *DeptApi) buildDeptPage(depts []models.SysDept) GetRep {
	deptMap := make(map[int]*Data)
	var deptTree GetRep

	for i := range depts {
		deptMap[depts[i].DeptId] = &Data{
			SysDept: &depts[i],
		}
	}

	for _, dept := range deptMap {
		dept.StatusDescription = parseStatus(dept.Status)
		// 如果父部门ID不存在于map中，视为顶级部门
		if parent, exists := deptMap[*dept.ParentId]; exists {
			parent.Children = append(parent.Children, dept)
		} else {
			deptTree = append(deptTree, dept)
		}
	}

	return deptTree
}

func parseStatus(status int) string {
	switch status {
	case 1:
		return "停用"
	case 2:
		return "正常"
	default:
		return ""
	}
}
