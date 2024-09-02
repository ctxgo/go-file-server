package dept

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GetTreeRep []*GetTreeData

type GetTreeData struct {
	Id       int            `json:"id"`
	Label    string         `json:"label"`
	Children []*GetTreeData `json:"children,omitempty"`
	ParentId int            `json:"-"`
}

func (api *DeptApi) GetTree(c *gin.Context) {
	data, err := api.makeTree(c)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(data).SendGin(c)

}

func (api *DeptApi) makeTree(c *gin.Context) (GetTreeRep, error) {
	list, err := api.deptRepo.Find(core.WithDeptDbScope(c))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data := api.buildDeptTree(list)
	return data, nil
}

func (api *DeptApi) buildDeptTree(depts []models.SysDept) GetTreeRep {
	deptMap := make(map[int]*GetTreeData)
	var deptTree GetTreeRep
	for _, dept := range depts {
		deptMap[dept.DeptId] = buildTreeData(dept)
	}

	for _, dept := range deptMap {
		if parent, exists := deptMap[dept.ParentId]; exists {
			parent.Children = append(parent.Children, dept)
		} else {
			deptTree = append(deptTree, dept)
		}
	}

	return deptTree
}

func buildTreeData(dept models.SysDept) *GetTreeData {
	return &GetTreeData{
		ParentId: *dept.ParentId,
		Id:       dept.DeptId,
		Label:    dept.DeptName,
	}
}
