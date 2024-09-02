package routers

import (
	"go-file-server/internal/services/normal/apis/auth"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r gin.IRouter, authApi *auth.Authenticator, captchaApi *auth.Captcha) {
	{
		r.GET("captcha", captchaApi.GenerateCaptchaHandler)
		r.POST("login", authApi.AuthHandler)
		r.POST("logout", authApi.Logout)
	}
}
