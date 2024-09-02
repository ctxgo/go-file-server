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

type Authenticator struct {
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	loginLogRepo *repository.LoginLogRepository
}

func NewAuthenticatorApi(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	loginLogRepo *repository.LoginLogRepository,
) *Authenticator {
	return &Authenticator{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		loginLogRepo: loginLogRepo,
	}
}

func (u *Authenticator) AuthHandler(c *gin.Context) {

	var loginVals LoginReq
	var status = "1"
	var msg = "登录成功"
	var userName = ""
	err := c.ShouldBind(&loginVals)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			status = "0"
			msg = "登录失败"
			c.Error(err)
		}
		go u.RecordLogin(c, userName, status, msg)
	}()
	var user *models.SysUser
	user, err = u.verify(loginVals)
	if err != nil {
		err = core.NewApiBizErr(err).SetMsg(err.Error()).SetBizCode(global.BizBadRequest)
		return
	}
	userName = user.Username
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

// Verify 校验验证码
func captchaVerify(id, code string, clear bool) bool {
	return base64Captcha.DefaultMemStore.Verify(id, code, clear)
}

func (u *Authenticator) verify(loginVals LoginReq) (*models.SysUser, error) {

	if !captchaVerify(loginVals.UUID, loginVals.Code, true) {

		return nil, errors.Errorf(global.ErrInvalidVerificationode)

	}
	user, err := u.userRepo.FindOne(
		repository.WithUsername(loginVals.Username),
		repository.WithStatus("2"))
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
func (u *Authenticator) createToken(user *models.SysUser) (string, string, error) {

	role, err := u.roleRepo.FindOne(repository.WithRoleId(user.RoleId))
	if err != nil {
		zlog.SugLog.Errorf("get role error, %s", err.Error())
		if err == gorm.ErrRecordNotFound {
			return "", "", core.NewApiBizErr(err).
				SetBizCode(global.BizAccessDenied).
				SetMsg("该用户无任何菜单权限")
		}
		return "", "", errors.Errorf(global.ErrServerNotOK)
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
