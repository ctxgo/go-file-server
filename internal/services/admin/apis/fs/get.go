package fs

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/internal/services/admin/apis/role"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/pathtool"
	"go-file-server/pkgs/utils/concurrentpool"
	"go-file-server/pkgs/utils/str"
	"go-file-server/pkgs/zlog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"golang.org/x/sync/errgroup"
)

type Item struct {
	Name    string `json:"name"`
	Mtime   string `json:"mtime"`
	RoleDir string `json:"roleDir"`
	Type    string `json:"type"`
	Size    string `json:"size"`
}

type GetPageRep struct {
	types.Page
	Items       []Item   `json:"items"`
	Permissions []string `json:"permissions"`
}

type GetReq struct {
	Name    string `form:"name"`
	OnlyDir bool   `form:"onlyDir"`
	Rid     string `form:"rid"`
	Action  string `form:"action" binding:"oneof=list download"`
	utils.UriPath
	types.Pagination
}

func (api *FsApi) GetPage(c *gin.Context) {
	var req GetReq
	err := core.ShouldBinds(c, &req, core.BindQuery, core.BindUri)
	if err != nil {
		c.Error(err)
		return
	}
	roleKey := core.ExtractClaims(c).RoleKey

	permissionsResults := make(chan []string, 1)
	getPageResults := make(chan GetPageRep, 1)
	var g errgroup.Group
	g.Go(func() error {
		permissions := api.getPermissions(roleKey, c.Request.URL.Path)
		permissionsResults <- permissions
		return nil
	})
	g.Go(func() error {
		data, err := api.getPage(c, req)
		if err != nil {
			return err
		}
		getPageResults <- data
		return nil
	})
	err = g.Wait()
	close(permissionsResults)
	close(getPageResults)
	if err != nil {
		c.Error(err)
		return
	}
	permissions := <-permissionsResults
	data := <-getPageResults
	data.Permissions = permissions
	data.PageIndex = req.PageIndex
	if len(data.Items) == 0 {
		data.Items = []Item{}
	}
	core.OKRep(data).SendGin(c)
}

func (api *FsApi) getPermissions(roleKey, path string) []string {

	if roleKey == models.AdminRoleKey {
		return []string{"POST", "DELETE", "PUT", "GET"}

	}

	data := []string{}
	policies := api.casbinEnforcer.GetFilteredPolicy(0, roleKey, "", "", "fs")

	for _, p := range policies {
		rolePath := strings.TrimSuffix(p[1], ".*")

		if !strings.HasPrefix(path, rolePath) {
			continue
		}
		if funk.Contains(data, p[2]) {
			continue
		}
		data = append(data, p[2])

	}
	return data

}

func (api *FsApi) getPage(c *gin.Context, getReq GetReq) (GetPageRep, error) {
	roleKey := core.ExtractClaims(c).RoleKey

	if strings.HasPrefix(getReq.Path, "/.tmp") {
		_, err := api.ensureTempDir(roleKey)
		if err != nil {
			return GetPageRep{}, err
		}
	}

	realPath, err := utils.GetRealPath(getReq.Path)
	if err != nil {
		return GetPageRep{}, core.NewApiBizErr(err).
			SetMsg(err.Error())
	}
	if roleKey == models.AdminRoleKey {
		return api.listAdminPath(realPath, getReq)
	}

	return api.listNormalPath(roleKey, realPath, getReq)
}

func (api *FsApi) listNormalPath(roleKey, realPath string, getReq GetReq) (GetPageRep, error) {
	var data GetPageRep
	if getReq.Path != "/" {
		roleDir, err := parseRoleDir(getReq.Rid)
		if err != nil {
			return data, errors.WithStack(err)
		}
		getReq.Rid = roleDir
		return api.listPath(realPath, getReq)
	}
	policies := api.casbinEnforcer.GetFilteredPolicy(0, roleKey, "", "GET", "fs")
	if len(policies) == 0 {
		return data, core.NewApiBizErr(nil).SetMsg("无任何目录权限，请联系管理员赋权")
	}
	for _, p := range policies {

		fsPath := role.ParseFsRolepath(p[1])
		if fsPath == "/" {
			return api.listPath(realPath, getReq)
		}

		realPath, err := utils.GetRealPath(fsPath)
		if err != nil {
			return data, err
		}

		filesDetails := pathtool.NewFiletool(realPath).GetFsDetails()
		if err := filesDetails.Err; err != nil {
			zlog.SugLog.Error(err)
			continue
		}
		if getReq.OnlyDir && filesDetails.Type != "dir" {
			continue
		}
		data.Count += 1
		items := makeItem(fsPath, filesDetails)
		items.Name = fsPath
		data.Items = append(
			data.Items,
			items,
		)
	}
	data.PageSize = int(data.Count)
	return data, nil

}

func (api *FsApi) listAdminPath(realPath string, getReq GetReq) (GetPageRep, error) {
	if getReq.OnlyDir {
		return api.listPathOnlyDir(realPath)
	}
	return api.listPath(realPath, getReq)
}

func makeItem(roledir string, f pathtool.FilesDetails) Item {
	return Item{
		Name:    f.Name,
		Mtime:   f.ModTime.Format(time.DateTime),
		Type:    f.Type,
		Size:    core.FormatBytes(uint64(f.Size)),
		RoleDir: roledir,
	}
}

func (api *FsApi) find(path string, req GetReq) ([]repository.FileDocument, uint64, error) {
	querys := []repository.FsScope{
		repository.WithPagination(req.PageIndex, req.PageSize),
	}

	if req.Name != "" {
		querys = append(querys,
			repository.WithParentPathPrefix(path),
			repository.WithRegexpName(".*"+req.Name+".*"))

	} else {
		querys = append(querys,
			repository.WithTermParentPath(path),
		)

	}
	return api.fsRepo.Find(querys...)

}

func (api *FsApi) listPathOnlyDir(path string) (GetPageRep, error) {
	var data GetPageRep
	data.Items = []Item{}
	filesList, _, err := api.fsRepo.Find(
		repository.WithTermParentPath(path),
		repository.WithIsDir(true),
	)
	if err != nil {
		return data, errors.WithStack(err)
	}
	for _, f := range filesList {
		data.Items = append(data.Items, Item{Name: f.Name})
	}
	return data, nil
}

func parseRoleDir(rid string) (string, error) {
	if rid == "" {
		return "", nil
	}
	return str.DecodeBase64(rid)
}

func (api *FsApi) listPath(realPath string, getReq GetReq) (GetPageRep, error) {
	var data GetPageRep

	filesList, total, err := api.find(realPath, getReq)
	if err != nil {
		return data, errors.WithStack(err)
	}
	pool, err := concurrentpool.NewAntsPool()
	defer pool.Release()
	if err != nil {
		return data, errors.WithStack(err)
	}
	itemsChan := make(chan Item, len(filesList))
	for _, f := range filesList {
		_f := f
		pool.Submit(
			func() {
				item, err := processFile(_f.Path, realPath, getReq)
				if err != nil {
					return
				}
				itemsChan <- item
			},
		)

	}
	pool.Wait()
	close(itemsChan)
	for item := range itemsChan {
		data.Items = append(data.Items, item)
	}
	data.Count = int64(total)
	data.PageSize = len(data.Items)
	return data, nil
}

func processFile(filePath, findPath string, getReq GetReq) (Item, error) {
	var item Item
	details := pathtool.NewFiletool(filePath).GetFsDetails()
	if details.Err != nil {
		zlog.SugLog.Error(details.Err)
		return Item{}, details.Err
	}

	item = makeItem(getReq.Rid, details)

	if getReq.Name == "" {
		return item, nil
	}

	normalizedPrefix := findPath
	if normalizedPrefix != "/" {
		normalizedPrefix += "/"
	}

	item.Name = strings.TrimPrefix(details.Path, normalizedPrefix)
	return item, nil
}
