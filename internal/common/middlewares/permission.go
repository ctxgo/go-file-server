package middlewares

import (
	"fmt"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/models"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// AuthCheckRole 权限检查中间件
func AuthCheckRole(svc *types.SvcCtx) gin.HandlerFunc {

	return func(c *gin.Context) {

		err := HandlerCheckRole(c, svc.CasbinEnforcer)
		if err != nil {
			core.HandlingErr(c, err)
		}
	}
}

func HandlerCheckRole(c *gin.Context, casbinEnforcer *casbin.CachedEnforcer) error {
	jwtClaims := core.ExtractClaims(c)
	//检查权限
	if jwtClaims.RoleKey == models.AdminRoleKey {
		return nil
	}
	err := CasbinEnforce(casbinEnforcer, jwtClaims.RoleKey, c.Request.URL.Path, c.Request.Method)
	if err != nil {
		return core.NewApiErr(err).
			SetHttpCode(global.HttpSuccess).
			SetBizCode(global.BizAccessDenied)
	}
	return nil
}

func CasbinEnforce(casbinEnforcer *casbin.CachedEnforcer, roleKey, path, method string) error {
	res, err := casbinEnforcer.Enforce(roleKey, path, method)
	if err != nil {
		return fmt.Errorf("AuthCheckRole error:%s method:%s path:%s", err, method, path)
	}
	if res {
		return nil
	}
	return fmt.Errorf("isTrue: %v role: %s method: %s path: %s message: %s", res, roleKey, method, path, "当前request无权限，请管理员确认！")
}
