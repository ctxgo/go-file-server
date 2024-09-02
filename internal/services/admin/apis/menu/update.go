package menu

import (
	"go-file-server/internal/common/core"
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
)

type UpdateRep struct {
	DeptId int `deptId:"id" binding:"required"`
}
type UpdateReq struct {
	MenuId int `uri:"id" comment:"编码"` // 编码
	CreateReq
}

func (api *MenuApi) UpdateDept(c *gin.Context) {
	var updateReq UpdateReq
	err := c.ShouldBind(&updateReq)
	if err != nil {
		c.Error(err)
		return
	}
	deptId, err := api.updateDept(c, updateReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(
		UpdateRep{
			DeptId: deptId,
		}).
		SendGin(c)
}
func (api *MenuApi) updateDept(c *gin.Context, updateReq UpdateReq) (int, error) {
	claims := core.ExtractClaims(c)

	sysMenu := &models.SysMenu{
		MenuId:     updateReq.MenuId,
		MenuName:   updateReq.MenuName,
		Title:      updateReq.Title,
		Path:       updateReq.Path,
		MenuType:   updateReq.MenuType,
		Action:     updateReq.Action,
		Permission: updateReq.Permission,
		ControlBy: coreModels.ControlBy{
			UpdateBy: claims.UserId,
		},
	}

	err := api.menuRepo.Save(sysMenu)
	return sysMenu.MenuId, err

}
