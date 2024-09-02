package dept

import (
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/services/admin/models"

	"go-file-server/internal/common/core"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateDeptReq struct {
	DeptId   int    `json:"deptId" binding:"required"`
	DeptPath string `json:"deptPath" binding:"required"`
	CreateReq
	coreModels.ControlBy
	coreModels.ModelTime
}

func (api *DeptApi) UpdateDept(c *gin.Context) {
	var updateDeptReq UpdateDeptReq
	err := c.ShouldBind(&updateDeptReq)
	if err != nil {
		c.Error(err)
		return
	}
	err = api.updateDept(c, updateDeptReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).
		SendGin(c)
}
func (api *DeptApi) updateDept(c *gin.Context, updateDeptReq UpdateDeptReq) error {
	claims := core.ExtractClaims(c)

	data := models.SysDept{
		DeptId:    updateDeptReq.DeptId,
		ParentId:  updateDeptReq.ParentId,
		DeptName:  updateDeptReq.DeptName,
		DeptPath:  updateDeptReq.DeptPath,
		Sort:      updateDeptReq.Sort,
		Leader:    updateDeptReq.Leader,
		Phone:     updateDeptReq.Phone,
		Email:     updateDeptReq.Email,
		Status:    updateDeptReq.Status,
		ControlBy: updateDeptReq.ControlBy,
		ModelTime: updateDeptReq.ModelTime,
	}
	data.UpdateBy = claims.UserId
	data.UpdatedAt = time.Now()
	return api.deptRepo.Save(&data)
}
