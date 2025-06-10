package fs

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (api *FsApi) rename(c *gin.Context, req UpdateReq) {

	err := middlewares.HandlerCheckRole(c, api.casbinEnforcer)
	if err != nil {
		c.Error(err)
		return
	}

	if err := utils.CheckFsName(req.NewName); err != nil {
		core.ErrBizRep().
			SetMsg(err.Error()).
			SendGin(c)
		return
	}
	realPath, err := utils.GetRealPath(req.Path)
	if err != nil {
		core.ErrBizRep().
			SetMsg(err.Error()).
			SendGin(c)
		return
	}

	newPath := filepath.Join(filepath.Dir(realPath), req.NewName)
	err = api.execRename(realPath, newPath)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(nil).SendGin(c)
}
