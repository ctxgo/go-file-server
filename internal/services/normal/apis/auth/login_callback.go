package auth

import (
	"context"
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/config"
	"strconv"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Claims struct {
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	PreferredUsername string   `json:"preferred_username"`
	Groups            []string `json:"groups,omitempty"` // 如果有组信息
}

func (u *Authenticator) LoginCallback(c *gin.Context) {
	ctx := c.Request.Context()

	if errMsg := c.Query("error"); errMsg != "" {
		core.ErrBizRep().SetMsg(errMsg)
		return
	}

	if state := c.Query("state"); state != config.OAuthCfg.State {
		core.ErrBizRep().SetMsg("无效的state参数")
		return
	}

	code := c.Query("code")
	if code == "" {
		core.ErrBizRep().SetMsg("请求中没有 code")
		return
	}

	claims, err := verifyAndDecode(ctx, code)
	if err != nil {
		c.Error(err)
		return
	}
	user, err := u.syncUser(claims)
	if err != nil {
		c.Error(err)
		return
	}
	token, expire, err := u.createToken(user)
	if err != nil {
		c.Error(err)
		return
	}
	core.OKRep(
		LoginRep{
			Token:        token,
			RefreshToken: token,
			Expire:       expire},
	).SendGin(c)
}

func verifyAndDecode(ctx context.Context, code string) (*Claims, error) {
	oauth2Config, provider, err := getOauthConfig(ctx)
	if err != nil {
		return nil, err
	}
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, core.NewApiBizErr(err).SetMsg("无法交换令牌")
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, core.NewApiBizErr(err).SetMsg("令牌响应中没有 id_token 字段")
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: config.OAuthCfg.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, core.NewApiBizErr(err).SetMsg("无法验证 ID Token")
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, core.NewApiBizErr(err).SetMsg("无法解析 ID Token 声明")
	}

	return &claims, err
}

func (u *Authenticator) syncUser(claims *Claims) (*models.SysUser, error) {
	data, err := u.userRepo.FindOne(repository.WithUsername(claims.Name), repository.WithPreloadDept())
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return u.createUserAndGroup(claims)
	}

	if data.Email == claims.Email && data.Source == "ldap" {
		return data, nil
	}

	err = u.userRepo.Update(func(su *models.SysUser) {
		su.Email = claims.Email
		su.Source = "ldap"
		su.Remark = "update for ldap"
	}, repository.WithUserId(data.UserId))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *Authenticator) createUserAndGroup(claims *Claims) (*models.SysUser, error) {
	var sysDept models.SysDept
	var user models.SysUser
	var groupName string = "default"
	if len(claims.Groups) != 0 {
		groupName = claims.Groups[0]
	}

	err := u.deptRepo.Repo.WithTransaction(
		func(tx *gorm.DB) (err error) {
			deptTsRepo := repository.NewDeptRepository(tx)

			sysDept, err = deptTsRepo.FindOne(repository.WithDeptName(groupName))
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
				sysDept = models.SysDept{
					ParentId: core.GetIntPointer(0),
					DeptName: groupName,
					Status:   2,
				}
				if err := deptTsRepo.Create(&sysDept); err != nil {
					return err
				}
				err = deptTsRepo.Update(func(sd *models.SysDept) {
					sd.DeptPath = "/0/" + strconv.Itoa(sysDept.DeptId) + "/"
				}, repository.WithByDeptId(sysDept.DeptId))
				if err != nil {
					return err
				}
			}
			user = models.SysUser{
				Username: claims.Name,
				NickName: claims.Name,
				Status:   "2",
				Email:    claims.Email,
				DeptId:   sysDept.DeptId,
				Remark:   "form ldap",
				Source:   "ldap",
			}
			return repository.NewUserRepository(tx).Create(&user)

		},
	)
	return &user, err

}
