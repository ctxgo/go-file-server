package auth

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/common/types"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/zlog"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type LoginReq struct {
	Username string `form:"UserName" json:"username" binding:"required"`
	Password string `form:"Password" json:"password" binding:"required"`
	Code     string `form:"Code" json:"code" binding:"required"`
	UUID     string `form:"UUID" json:"uuid" binding:"required"`
}

type LoginRep struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Expire       string `json:"expire"`
}

func (u *Authenticator) AuthHandler(c *gin.Context) {

	var loginVals LoginReq
	var status = "1"
	var msg = "web"
	err := c.ShouldBind(&loginVals)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			status = "2"
			msg = "web"
			c.Error(err)
		}
		go u.RecordLogin(c, loginVals.Username, status, msg)
	}()
	var user *models.SysUser
	user, err = u.verify(loginVals)
	if err != nil {
		err = core.NewApiBizErr(err).SetMsg(err.Error()).SetBizCode(global.BizBadRequest)
		return
	}
	var token, expire string
	token, expire, err = u.createToken(user)
	if err != nil {
		return
	}
	core.OKRep(
		LoginRep{
			Token:        token,
			RefreshToken: token,
			Expire:       expire},
	).SendGin(c)
}

func (u *Authenticator) verify(loginVals LoginReq) (*models.SysUser, error) {

	if !VerifyCaptcha(loginVals.UUID, loginVals.Code, true) {

		return nil, errors.Errorf(global.ErrInvalidVerificationode)

	}
	return VerifyUser(u.userRepo, loginVals)
}

func (u *Authenticator) createToken(user *models.SysUser) (string, string, error) {
	return CreateToken(u.roleRepo, user)
}

// Verify 校验验证码
func VerifyCaptcha(id, code string, clear bool) bool {
	return base64Captcha.DefaultMemStore.Verify(id, code, clear)
}

func VerifyUser(userRepo *repository.UserRepository, loginVals LoginReq) (*models.SysUser, error) {

	user, err := userRepo.FindOne(
		repository.WithUsername(loginVals.Username),
		repository.WithUserStatus("2"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Errorf(global.ErrFailedAuthentication)
		}
		return nil, errors.Errorf(global.ErrServerNotOK)
	}
	_, err = core.CompareHashAndPassword(user.Password, loginVals.Password)
	if err != nil {
		zlog.SugLog.Errorf("user login error, %s", err.Error())
		return nil, errors.Errorf(global.ErrFailedAuthentication)
	}
	return user, nil
}

// CreateToken 生成一个token
func CreateToken(roleRepo *repository.RoleRepository, user *models.SysUser) (string, string, error) {
	role, err := roleRepo.FindOne(repository.WithRoleId(user.RoleId))
	if err != nil {
		zlog.SugLog.Errorf("get role error, %s", err.Error())
		if err != gorm.ErrRecordNotFound {
			return "", "", errors.Errorf(global.ErrServerNotOK)
		}
		role = &models.SysRole{}
	}

	return middlewares.CreateToken(
		func(jc *types.JwtClaims) {
			jc.RoleId = role.RoleId
			jc.RoleKey = role.RoleKey
			jc.RoleName = role.RoleName
			jc.Username = user.Username
			jc.UserId = user.UserId
		},
	)
}

func (u *Authenticator) RecordLogin(c *gin.Context, userName, status, msg string) {
	ua := core.GetUserAgent(c)

	u.loginLogRepo.Create(&models.SysLoginLog{
		Username: userName,
		Status:   status,
		Ipaddr:   core.GetClientIP(c),
		Browser:  core.FormatBrowserInfo(ua),
		Os:       ua.OS(),
		Platform: ua.Platform(),
		Msg:      msg,
	})

}
