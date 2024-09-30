package core

import (
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/types"

	"github.com/gin-gonic/gin"
)

func ExtractClaims(c *gin.Context) *types.JwtClaims {
	claims, exists := c.Get(global.JwtPayloadKey)
	if !exists {
		return &types.JwtClaims{}
	}
	return claims.(*types.JwtClaims)
}

func GetUserId(c *gin.Context) int {
	claimsData := ExtractClaims(c)
	return claimsData.UserId
}

func GetRoleKey(c *gin.Context) string {
	claimsData := ExtractClaims(c)
	return claimsData.RoleKey
}

func GetUserName(c *gin.Context) string {
	claimsData := ExtractClaims(c)
	return claimsData.Username
}
