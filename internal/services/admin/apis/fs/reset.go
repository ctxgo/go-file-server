package fs

import (
	"go-file-server/internal/common/core"

	"github.com/gin-gonic/gin"
)

func (api *FsApi) Reset(c *gin.Context) {

	if err := core.AssertAdmin(c); err != nil {
		c.Error(err)
		return
	}

	if err := api.fsRepo.ResetIndex(); err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)
}
