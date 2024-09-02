package fs

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/services/admin/apis/fs/utils"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	utils.UriPath
	NewName string `json:"newName"`
}

func (api *FsApi) Create(c *gin.Context) {
	switch contentType := c.ContentType(); contentType {
	case "application/json":
		api.createDir(c)

	case "multipart/form-data":
		api.upLoad(c)

	default:
		core.ErrRep().
			SetHttpCode(global.BadRequestError).
			SetMsg("Unsupported Content-Type: " + contentType).
			SendGin(c)
	}
}

func (api *FsApi) createDir(c *gin.Context) {
	var req CreateReq
	err := core.ShouldBinds(c, &req, core.BindJson, core.BindUri)
	if err != nil {
		c.Error(err)
		return
	}
	if err := utils.CheckFsName(req.NewName); err != nil {
		core.ErrBizRep().SetMsg(err.Error()).SendGin(c)
		return
	}

	realPath, err := utils.GetRealPath(req.Path, req.NewName)
	if err != nil {
		core.ErrBizRep().SetMsg(err.Error()).SendGin(c)

		return
	}

	err = api.fsRepo.CreateDir(realPath)
	if err != nil {
		if ok, err := utils.ParsePathErr(err); ok {
			core.ErrBizRep().SetMsg(err.Error()).SendGin(c)
			return
		}
		c.Error(err)
		return
	}
	core.OKRep(nil).SendGin(c)
}

func (api *FsApi) upLoad(c *gin.Context) {
	var req CreateReq
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.Error(err)
		return
	}
	filepart, err := c.FormFile("file")
	if err != nil {
		core.ErrRep().SetHttpCode(global.BadRequestError).SendGin(c)
		return
	}

	filePath, err := utils.GetRealPath(req.Path, filepart.Filename)
	if err != nil {
		core.ErrRep().SetHttpCode(global.BadRequestError).SendGin(c)
	}
	err = c.SaveUploadedFile(filepart, filePath)
	if err != nil {
		core.ErrRep().SendGin(c)
		return
	}
	err = api.fsRepo.AddResource(filePath)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(nil).SendGin(c)

}
