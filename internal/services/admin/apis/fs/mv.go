package fs

import (
	"fmt"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/internal/services/admin/models"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (api *FsApi) move(c *gin.Context, req UpdateReq) {

	if err := validateMoveFiled(req); err != nil {
		core.ErrBizRep().
			SetMsg(err.Error()).
			SendGin(c)
		return
	}
	if req.Path == req.Destination {
		core.ErrBizRep().SetMsg("源路径和目标路径相同").SendGin(c)
		return
	}

	if err := api.execMove(c, req); err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)
}
func (api *FsApi) checMovePermission(c *gin.Context, req UpdateReq) error {
	roleKey := core.ExtractClaims(c).RoleKey
	if roleKey == models.AdminRoleKey {
		return nil
	}
	ok, err := api.casbinEnforcer.Enforce(
		roleKey,
		c.Request.URL.Path,
		"DELETE",
	)
	if err != nil {
		return err
	}

	if !ok {
		err = errors.Errorf("role: %s , path:%s, 无删除权限", roleKey, c.Request.URL.Path)
		return core.NewApiBizErr(err).SetBizCode(global.BizAccessDenied).
			SetMsg(fmt.Sprintf("您没有当前目录 %s 的删除权限", req.Path))
	}

	desApi, err := utils.SafeJoinPath("/api/v1/fs", req.Destination)

	if err != nil {
		return err
	}

	ok, err = api.casbinEnforcer.Enforce(
		roleKey,
		desApi,
		"CREATE",
	)

	if err != nil {
		return err
	}

	if !ok {
		err = errors.Errorf("role: %s , path:%s, 无创建权限", roleKey, desApi)

		return core.NewApiBizErr(err).SetBizCode(global.BizAccessDenied).
			SetMsg(fmt.Sprintf("您没有目标目录 %s 创建资源的权限", req.Destination))
	}
	return nil

}

func (api *FsApi) execMove(c *gin.Context, req UpdateReq) error {
	err := api.checMovePermission(c, req)
	if err != nil {
		return err
	}

	realPath, err := utils.GetRealPath(req.Path)
	if err != nil {
		return core.NewApiBizErr(err).SetMsg(err.Error())
	}

	destination, err := utils.GetRealPath(req.Destination, filepath.Base(realPath))
	if err != nil {
		return core.NewApiBizErr(err).SetMsg(err.Error())

	}

	return api.execRename(realPath, destination)

}

func validateMoveFiled(req UpdateReq) error {
	if req.Destination == "" {
		return errors.Errorf("action 为 move 时 ,destination(目标路径)字段不能为空")
	}
	return nil
}
