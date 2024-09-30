package fs

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/pkgs/zlog"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (api *FsApi) Delete(c *gin.Context) {
	var req utils.UriPath
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.Error(err)
		return
	}

	roleKey := core.ExtractClaims(c).RoleKey
	err = api.execDelete(req.Path, roleKey)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(nil).SendGin(c)
}

func (api *FsApi) execDelete(path string, roleKey string) (err error) {

	handleErr := func(err error) error {
		return core.NewApiBizErr(err).SetMsg(err.Error())
	}

	srcPath, err := utils.GetRealPath(path)
	if err != nil {
		return handleErr(err)
	}

	// 直接删除
	if strings.HasPrefix(srcPath, utils.GetTmpDir()) {
		return api.fsRepo.RemoveAll(srcPath)
	}

	tmpDir, err := api.ensureTempDir(roleKey)
	if err != nil {
		return err
	}

	// 转移到回收站
	desPath := filepath.Join(tmpDir, filepath.Base(path)+"_"+utils.GetTimeStr())

	err = api.fsRepo.Rename(srcPath, desPath)
	if err != nil {
		ok, err := utils.ParsePathErr(err)
		if ok {
			return core.NewApiBizErr(err).
				SetMsg(err.Error())
		}
		zlog.SugLog.Error(err)
		return errors.Errorf("内部异常")
	}
	return nil
}
