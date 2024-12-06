package role

import (
	"fmt"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/base"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type UpdateFsRep struct {
	RoleId int `json:"roleId" comment:"角色编码"` // 角色编码
}

type FsPermissions struct {
	Path        string   `json:"path" binding:"required"`
	Permissions []string `json:"permissions" binding:"required,min=1,dive,oneof=POST DELETE PUT GET"`
}

type UpdateFsReq struct {
	RoleId        int             `json:"roleId" binding:"required"`
	RateLimit     uint64          `json:"rateLimit"`
	FsPermissions []FsPermissions `json:"fsRoles"`
}

const RateLimitKey = "RateLimitKey"

func (api *RoleApi) UpdateFs(c *gin.Context) {
	var updateReq UpdateFsReq
	err := c.ShouldBind(&updateReq)
	if err != nil {
		c.Error(err)
		return
	}
	err = api.updateFs(updateReq)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(nil).SendGin(c)
}

func (api *RoleApi) updateFs(updateReq UpdateFsReq) error {
	var g errgroup.Group
	g.Go(func() error {
		return api.updateFsRateLimit(updateReq)
	})
	g.Go(func() error {
		return api.updateFsPermissions(updateReq)
	})
	return g.Wait()
}

func (api *RoleApi) updateFsRateLimit(updateReq UpdateFsReq) error {

	err := api.roleRepo.Update(func(sr *models.SysRole) {
		sr.RateLimit = updateReq.RateLimit
	}, repository.WithRoleId(updateReq.RoleId),
		base.WithSelect("rate_limit"),
	)
	if err != nil {
		return err
	}
	return api.cache.Set(
		fmt.Sprintf("%d-%s", updateReq.RoleId, RateLimitKey),
		updateReq.RateLimit, 0)
}

func buildRolePath(path string) string {
	if path == "/" {
		return "/api/v1/fs/.*"
	}
	return filepath.Join("/api/v1/fs", path) + ".*"

}

func (api *RoleApi) updateFsPermissions(updateReq UpdateFsReq) error {

	role, err := api.roleRepo.FindOne(
		repository.WithRoleId(updateReq.RoleId),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(updateReq.FsPermissions) != 0 {
		addDefaultRole(role.RoleKey, &updateReq)
	}

	var policiesToRemove [][]string
	var policiesToAdd [][]string
	mp := make(map[string][]string, 0)
	for _, p := range updateReq.FsPermissions {
		path := p.Path
		path = buildRolePath(path)
		for _, action := range p.Permissions {
			mp[role.RoleKey+"-"+path+"-"+action+"-fs"] = []string{
				role.RoleKey, path, action, "fs"}
		}
	}
	policies := api.casbinEnforcer.GetFilteredPolicy(0, role.RoleKey, "", "", "fs")

	for _, p := range policies {

		key := strings.Join(p, "-")
		_, ok := mp[key]
		if ok {
			delete(mp, key)
		} else {
			policiesToRemove = append(policiesToRemove, p)
		}

	}
	for _, v := range mp {
		policiesToAdd = append(policiesToAdd, v)
	}

	if len(policiesToRemove) > 0 {
		_, err := api.casbinEnforcer.RemovePolicies(policiesToRemove)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if len(policiesToAdd) > 0 {
		_, err = api.casbinEnforcer.AddNamedPolicies("p", policiesToAdd)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err = api.casbinEnforcer.LoadPolicy()
	return errors.WithStack(err)

}

func addDefaultRole(roleKey string, updateReq *UpdateFsReq) {
	updateReq.FsPermissions = append(updateReq.FsPermissions,
		FsPermissions{
			Path:        "/.tmp/" + roleKey,
			Permissions: []string{"POST", "DELETE", "PUT", "GET"},
		},
	)
}
