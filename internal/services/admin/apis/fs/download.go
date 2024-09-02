package fs

import (
	"encoding/json"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/internal/services/admin/apis/role"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/pathtool"
	"go-file-server/pkgs/utils/limiter"
	"go-file-server/pkgs/utils/str"
	"go-file-server/pkgs/utils/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type DownloadInfo struct {
	Path  string
	Token string
}

func (api *FsApi) GetDownloadUrl(c *gin.Context) {
	var req utils.UriPath
	err := core.ShouldBinds(c, &req, core.BindUri)
	if err != nil {
		c.Error(err)
		return
	}
	rolekey := core.ExtractClaims(c).RoleKey
	err = api.checkDownloadPermission(rolekey, req.Path)
	if err != nil {
		c.Error(err)
		return
	}
	token, err := middlewares.GetToken(c)
	if err != nil {
		c.Error(err)
		return
	}
	id, err := api.makeID(req.Path, token)
	if err != nil {
		c.Error(err)
		return
	}

	isDIr, realPath, err := checkPath(req.Path)
	if err != nil {
		return
	}
	fileName := filepath.Base(realPath)
	if isDIr {
		fileName += ".zip"
	}
	host := utils.GetHost(c)
	if host == "" {
		core.ErrBizRep().
			SetMsg("无法获取原始请求地址，请联系管理员设置转发请求头x-forwarded-host").
			SendGin(c)
		return
	}
	url := host + filepath.Join("/api/v1/fsd", id, fileName)
	core.OKRep(url).SendGin(c)
}

func (api *FsApi) makeID(uriPath, token string) (id string, err error) {
	var sdata string
	defer func() {
		if err != nil {
			return
		}
		err = api.cache.Set(id, sdata, 3*time.Hour)
	}()

	data := DownloadInfo{Path: uriPath, Token: token}

	sdata, err = str.ConvertToString(data)
	if err != nil {
		return "", err
	}
	id, ok := api.idManager.GetID(sdata)
	if ok {
		return id, nil
	}

	id, err = str.NextStrID()
	if err != nil {
		return "", err
	}
	id = api.idManager.GetOrCreateID(sdata, id)
	return id, nil
}

func (api *FsApi) Download(c *gin.Context) {
	var req utils.UriPath
	var err error
	var downloadInfo DownloadInfo
	defer func() {
		if err != nil {
			c.Error(err)
		}
	}()
	err = core.ShouldBinds(c, &req, core.BindUri)
	if err != nil {
		return
	}
	downloadInfo, err = api.pasreUri(req.Path)
	if err != nil {
		return
	}
	jwtClaims, err := api.parseToken(downloadInfo)
	if err != nil {
		return
	}
	err = api.checkDownloadPermission(jwtClaims.RoleKey, downloadInfo.Path)
	if err != nil {
		return
	}
	err = api.send(c, jwtClaims, downloadInfo.Path)
}

func (api *FsApi) checkDownloadPermission(roleKey, uriPath string) error {
	if roleKey == "admin" {
		return nil
	}
	apiPath := filepath.Join("/api/v1/fs/", uriPath)

	ok, err := api.casbinEnforcer.Enforce(
		roleKey,
		apiPath,
		"GET",
	)
	if err != nil {
		return core.NewApiErr(err)
	}
	if !ok {
		return core.NewApiErr(nil).
			SetHttpCode(global.UnauthorizedError).
			SetMsg("无权限")
	}
	return nil
}

func (api *FsApi) parseToken(downloadInfo DownloadInfo) (*types.JwtClaims, error) {
	jwtClaims, err := middlewares.ParseToken(downloadInfo.Token)
	if err != nil {
		return nil, core.NewApiErr(err).
			SetBizCode(global.BizUnauthorizedErr).
			SetMsg(err.Error())
	}
	return jwtClaims, err
}

func (api *FsApi) pasreUri(path string) (DownloadInfo, error) {
	var downloadInfo DownloadInfo
	urlPath := strings.TrimPrefix(path, "/")
	paths := strings.Split(urlPath, "/")
	if len(paths) != 2 {
		return downloadInfo, core.NewApiErr(nil).
			SetHttpCode(global.BadRequestError).
			SetMsg("路径解析失败")
	}
	id := paths[0]
	data, err := api.cache.Get(id)
	if err != nil {
		if cache.IsKeyNotFoundError(err) {
			return downloadInfo, core.NewApiBizErr(err).
				SetHttpCode(global.StatusNotFound).
				SetBizCode(global.BizNotFound).
				SetMsg("链接已失效")
		}
		return downloadInfo, core.NewApiErr(err)
	}

	err = json.Unmarshal([]byte(data), &downloadInfo)
	if err != nil {
		return downloadInfo, core.NewApiErr(err).
			SetHttpCode(global.BadRequestError).
			SetMsg("路径解析失败")
	}
	return downloadInfo, nil

}

func checkPath(path string) (isDir bool, realPath string, err error) {

	realPath, err = utils.GetRealPath(path)
	if err != nil {
		err = core.NewApiErr(err).
			SetHttpCode(global.BadRequestError).
			SetMsg(err.Error())
		return
	}
	isDir, err = pathtool.NewFiletool(realPath).AssertDir()
	if err != nil {
		if os.IsNotExist(err) {
			err = core.NewApiErr(err).
				SetHttpCode(global.StatusNotFound).
				SetBizCode(global.BizNotFound)
			return
		}
		err = core.NewApiErr(err)
	}
	return
}

func (api *FsApi) getRaleLimiteBytes(roleId int) (uint64, error) {
	rateLimitStr, err := api.cache.Get(fmt.Sprintf("%d-%s", roleId, role.RateLimitKey))
	if err == nil {
		return strconv.ParseUint(rateLimitStr, 10, 64)
	}
	if !cache.IsKeyNotFoundError(err) {
		return 0, err
	}
	data, err := api.roleRepo.FindOne(repository.WithRoleId(roleId))
	if err != nil {
		return 0, err
	}
	err = api.cache.Set(
		fmt.Sprintf("%d-%s", roleId, role.RateLimitKey),
		data.RateLimit,
		0,
	)

	return data.RateLimit, err
}

func (api *FsApi) send(c *gin.Context, jwtClaims *types.JwtClaims, path string) error {
	raleLimiteBytes, err := api.getRaleLimiteBytes(jwtClaims.RoleId)
	if err != nil {
		return err
	}
	raleLimiter := api.limiterManager.GetLimiter(jwtClaims.RoleKey, raleLimiteBytes)
	isDIr, realPath, err := checkPath(path)
	if err != nil {
		return err
	}
	if isDIr {
		return sendDir(c, realPath, raleLimiter)
	}
	return sendFile(c, realPath, raleLimiter)
}

func sendFile(c *gin.Context, src string, limiter *limiter.Limiter) error {
	fileName := filepath.Base(src)
	c.Header("Content-Type", "application/octet-stream")
	//强制浏览器下载
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	//浏览器下载或预览
	c.Header("Content-Disposition", "inline;filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	fs, err := os.Open(src)
	if err != nil {
		return errors.WithStack(err)
	}
	defer fs.Close()
	writer := limiter.LimitWriter(c.Request.Context(), c.Writer)
	c.Writer.Flush()
	_, err = io.Copy(writer, fs)
	return errors.WithStack(err)
}

func sendDir(c *gin.Context, src string, limiter *limiter.Limiter) error {
	fileName := filepath.Base(src) + ".zip"
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/zip")
	writer := limiter.LimitWriter(c.Request.Context(), c.Writer)
	err := zip.NewStreamZip(writer).ZipWithCtx(c.Request.Context(), src)
	return errors.WithStack(err)
}
