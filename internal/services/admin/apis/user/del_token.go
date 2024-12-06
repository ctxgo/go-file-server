package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type DeleteTokenReq struct {
	Ids []int `json:"ids" binding:"required,min=1"`
}

func (api *UserAPI) DeleteToken(c *gin.Context) {
	var deleteReq DeleteTokenReq
	err := c.ShouldBind(&deleteReq)
	if err != nil {
		c.Error(err)
		return
	}

	claims := core.ExtractClaims(c)

	err = api.deleteToken(claims, deleteReq)
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(nil).SendGin(c)
}

func (api *UserAPI) deleteToken(claims *types.JwtClaims, deleteReq DeleteTokenReq) error {

	data, err := api.userTokenRepo.Find(repository.WithUserTokenUserId(claims.UserId),
		repository.WithUserTokenIds(deleteReq.Ids...))
	if err != nil {
		return errors.WithStack(err)
	}
	for _, token := range data {
		err = api.cache.Set(middlewares.GenPersonalTokenRevokedKey(token.Token), "false", 24*time.Hour)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	err = api.userTokenRepo.Delete(repository.WithUserTokenUserId(claims.UserId),
		repository.WithUserTokenIds(deleteReq.Ids...))
	return errors.WithStack(err)

}
