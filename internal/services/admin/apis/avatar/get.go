package avatar

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (api *AvatarAPI) Get(c *gin.Context) {
	claims := core.ExtractClaims(c)
	data, err := api.avatarRepo.FindOne(repository.WithUserId(claims.UserId))
	if err == nil {
		c.Writer.Header().Set("Content-Disposition", "inline")
		c.Data(http.StatusOK, "application/octet-stream", data.Data)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		core.ErrBizRep().SetBizCode(global.BizNotFound).SendGin(c)
		return
	}
	c.Error(errors.WithStack(err))
}
