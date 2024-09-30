package fs

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"io"
	"os"
	"path/filepath"

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

	err = api.fsRepo.MkdirAll(realPath, os.ModePerm)
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
	err = api.saveUploadedFile(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)

}

func (api *FsApi) saveUploadedFile(c *gin.Context, req CreateReq) error {

	filepart, err := c.FormFile("file")
	if err != nil {
		return core.NewApiErr(err).SetHttpCode(global.BadRequestError)
	}

	dst, err := utils.GetRealPath(req.Path, filepart.Filename)
	if err != nil {
		return core.NewApiErr(err).SetHttpCode(global.BadRequestError)
	}

	src, err := filepart.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	claims := core.ExtractClaims(c)
	raleLimiter, err := api.getLimiter(claims.UserId, claims.RoleKey)
	if err != nil {
		return err
	}
	reader := raleLimiter.LimitReader(c.Request.Context(), src)
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}

	return api.fsRepo.AddResource(dst)

}
