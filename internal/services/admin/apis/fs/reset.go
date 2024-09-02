package fs

import (
	"go-file-server/internal/common/core"

	"github.com/gin-gonic/gin"
)

func (api *FsApi) Reset(c *gin.Context) {

	roleKey := core.ExtractClaims(c).RoleKey
	if roleKey != "admin" {
		core.ErrBizRep().SetMsg("无操作权限")
		return
	}
	err := api.fsRepo.ResetIndex()
	if err != nil {
		c.Error(err)
	}
	core.OKRep(nil).SendGin(c)
}
