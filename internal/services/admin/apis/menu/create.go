package menu

import (
	"go-file-server/internal/common/core"
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/config"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CreateRep struct {
	DeptId int `json:"dept_id" `
}

type CreateReq struct {
	MenuName   string `form:"menuName" comment:"菜单name"` //菜单name
	Title      string `form:"title" comment:"显示名称"`      //显示名称
	Path       string `form:"path" comment:"路径"`         //路径
	Paths      string `form:"paths" comment:"id路径"`      //id路径
	Permission string `form:"permission" comment:"权限编码"` //权限编码
	Action     string `form:"action" comment:"动作"`       //动作
	MenuType   string `form:"menuType" comment:"菜单类型"`   //菜单类型
}

func (api *MenuApi) Create(c *gin.Context) {
	var createVals CreateReq
	err := c.ShouldBind(&createVals)
	if err != nil {
		c.Error(err)
		return
	}

	deptId, err := api.create(c, createVals)

	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(
		CreateRep{
			DeptId: deptId,
		},
	).SendGin(c)
}

func (api *MenuApi) create(c *gin.Context, createVals CreateReq) (int, error) {
	claims := core.ExtractClaims(c)

	sysMenu := &models.SysMenu{
		MenuName: createVals.MenuName,
		Title:    createVals.Title,
		Path:     createVals.Path,
		MenuType: createVals.MenuType,
		ControlBy: coreModels.ControlBy{
			CreateBy: claims.UserId,
		},
	}
	err := api.menuRepo.Repo.WithTransaction(
		func(tx *gorm.DB) error {
			menuTsRepo := repository.NewMenuRepository(tx)
			err := menuTsRepo.Create(sysMenu)
			if err != nil {
				return err
			}
			deptPath, err := api.makeDeptPath(sysMenu)
			if err != nil {
				return err
			}
			return menuTsRepo.Update(func(sd *models.SysMenu) {
				sd.Paths = deptPath
			}, repository.WithByDeptId(sysMenu.MenuId))
		},
	)
	if err != nil {
		if repository.IsDuplicateError(config.DatabaseCfg.Driver, err) {
			return 0, core.NewApiBizErr(errors.WithStack(err)).
				SetMsg("权限已经存在")
		}
		return 0, errors.WithStack(err)
	}

	return sysMenu.MenuId, nil
}

func (api *MenuApi) makeDeptPath(menu *models.SysMenu) (string, error) {
	path := strconv.Itoa(menu.MenuId) + "/"
	if menu.ParentId == 0 {
		return "/0/" + path, nil
	}
	menuP, err := api.menuRepo.FindOne(repository.WithMenuId(menu.MenuId))
	if err != nil {
		return "", errors.WithStack(err)
	}
	return menuP.Paths + path, nil

}
