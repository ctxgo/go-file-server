package role

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

type GetInfoRep *models.SysRole

type GetInfoReq struct {
	Id int `uri:"id"`
}

func (api *RoleApi) GetInfo(c *gin.Context) {

	var getInfoReq GetInfoReq
	err := c.ShouldBindUri(&getInfoReq)
	if err != nil {
		c.Error(err)
		return
	}

	data, err := api.getInfo(getInfoReq)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(data).SendGin(c)
}

func (api *RoleApi) getInfo(getInfoReq GetInfoReq) (GetInfoRep, error) {
	var getInfoRep GetInfoRep
	data, err := api.roleRepo.FindOne(
		repository.WithPreloadSysMenu(),
		repository.WithRoleId(getInfoReq.Id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return getInfoRep, core.NewApiBizErr(err).
				SetBizCode(global.BizNotFound)
		}
		return getInfoRep, core.NewApiErr(errors.WithStack(err))
	}

	for _, x := range data.SysMenu {
		data.MenuIds = append(data.MenuIds, x.MenuId)
	}
	data.SysMenu = []models.SysMenu{}
	data.FsRoles = api.getFsRoles(data.RoleKey)
	return GetInfoRep(data), nil

}

func (api *RoleApi) getFsRoles(roleKey string) []models.FsPermissions {

	policies := api.casbinEnforcer.GetFilteredPolicy(0, roleKey, "", "", "fs")

	pathActions := make(map[string][]string)

	// 使用funk来处理数据
	for _, record := range policies {
		path, action := record[1], record[2]
		fsPath := ParseFsRolepath(path)
		if !funk.Contains(pathActions[fsPath], action) {
			pathActions[fsPath] = append(pathActions[fsPath], action)
		}
	}

	// 构建结果列表
	results := []models.FsPermissions{}
	for path, actions := range pathActions {
		results = append(results, models.FsPermissions{Path: path, Permissions: actions})
	}

	return results
}

func ParseFsRolepath(path string) string {
	path = strings.TrimPrefix(path, "/api/v1/fs/")
	path = strings.TrimSuffix(path, ".*")
	if path == "" {
		path = "/"
	}
	return path

}
