package auth

import (
	"go-file-server/internal/common/core"

	"github.com/gin-gonic/gin"
)

func (u *Authenticator) Logout(c *gin.Context) {
	core.OKRep(nil).SendGin(c)
}
