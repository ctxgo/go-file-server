package fs

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/apis/fs/utils"

	"github.com/gin-gonic/gin"
)

type UpdateReq struct {
	utils.UriPath
	Action      string `json:"action" binding:"required,oneof=rename move"`
	NewName     string `json:"newName"`
	Destination string `json:"destination"`
}

func (api *FsApi) Update(c *gin.Context) {
	var req UpdateReq
	err := core.ShouldBinds(c, &req, core.BindJson, core.BindUri)
	if err != nil {
		c.Error(err)
		return
	}
	if req.Action == "rename" {
		api.rename(c, req)
		return
	}
	api.move(c, req)
}
