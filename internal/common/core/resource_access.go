package core

import (
	"go-file-server/internal/common/global"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
)

// AssertAdmin 断言当前用户必须是管理员
// 如果不是管理员则返回错误，是管理员则返回nil
func AssertAdmin(c *gin.Context) error {
	if GetRoleKey(c) == models.AdminRoleKey {
		return nil
	}
	return NewApiErr(nil).
		SetBizCode(global.BizAccessDenied)
}

// VerifyResourceOwner 验证当前用户是否为资源所有者
// 用于业务层验证用户是否有权操作特定资源
// resourceOwnerID: 资源所有者ID（整数类型）
// 返回nil表示验证通过，否则返回错误
func VerifyResourceOwner(c *gin.Context, resourceOwnerID int) error {
	// 管理员直接放行
	claims := ExtractClaims(c)
	if claims.RoleKey == models.AdminRoleKey {
		return nil
	}

	// 获取当前用户ID
	currentUserId := claims.UserId

	// 检查用户是否为资源所有者
	if resourceOwnerID == currentUserId {
		return nil
	}

	// 不是资源所有者，返回错误
	return NewApiErr(nil).
		SetBizCode(global.BizAccessDenied)
}
