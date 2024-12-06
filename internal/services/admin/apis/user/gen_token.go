package user

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/middlewares"
	coreModels "go-file-server/internal/common/models"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type GenTokenReq struct {
	UserID int `json:"user_id" binding:"required"` // 绑定 user_id，并声明为必填项
}

type GenTokenRep struct {
	Token string `json:"token"`
}

func (api *UserAPI) GenToken(c *gin.Context) {
	var rep GenTokenRep
	var err error
	defer func() {
		if err != nil {
			c.Error(err)
		}
	}()
	var query GenTokenReq

	err = c.ShouldBind(&query)
	if err != nil {
		return
	}

	claims := core.ExtractClaims(c)

	if query.UserID != claims.UserId {
		err = core.NewApiBizErr(errors.Errorf("生成失败: 只能生成自己的token"))
		return
	}

	err = api.genToken(claims, &rep)
	if err != nil {
		return
	}

	core.OKRep(rep).SendGin(c)
}

func (api *UserAPI) genToken(claims *types.JwtClaims, rep *GenTokenRep) error {
	ntime := time.Now()
	token, _, err := middlewares.CreateToken(
		func(jc *types.JwtClaims) {
			jc.RoleId = claims.RoleId
			jc.RoleKey = claims.RoleKey
			jc.RoleName = claims.RoleName
			jc.Username = claims.Username
			jc.UserId = claims.UserId
			jc.IssuedAt = ntime.Unix()
			jc.ExpiresAt = 0
			jc.IsPersonalToken = true
		},
	)
	if err != nil {
		return err
	}

	data := &models.UserToken{
		UserID:    claims.UserId,
		Token:     token,
		ModelTime: coreModels.ModelTime{CreatedAt: ntime},
	}

	err = api.userTokenRepo.Create(data)

	if err != nil {
		return err

	}

	rep.Token = token
	return nil
}
