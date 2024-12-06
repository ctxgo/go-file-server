package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"

	"github.com/gin-gonic/gin"
)

type GetTokenReq struct {
	UserID int `form:"user_id" binding:"required"` // 绑定 user_id，并声明为必填项
}

type GetTokenRep []models.UserToken

func (api *UserAPI) GetToken(c *gin.Context) {
	var rep GetTokenRep
	var err error
	defer func() {
		if err != nil {
			c.Error(err)
		}
	}()
	var query GetTokenReq

	err = c.ShouldBindQuery(&query)
	if err != nil {
		return
	}
	rep, err = api.getToken(query.UserID)
	if err != nil {
		return
	}
	core.OKRep(rep).SendGin(c)
}

func (api *UserAPI) getToken(userID int) (GetTokenRep, error) {
	return api.userTokenRepo.Find(repository.WithUserTokenUserId(userID))
}
