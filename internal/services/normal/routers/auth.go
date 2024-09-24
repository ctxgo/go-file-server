package routers

import (
	"go-file-server/internal/services/normal/apis/auth"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r gin.IRouter, authApi *auth.Authenticator) {
	{
		r.POST("login", authApi.AuthHandler)
		r.GET("login/dex", authApi.LoginDex)
		r.GET("login/callback", authApi.LoginCallback)
		r.POST("logout", authApi.Logout)
		r.POST("oauthlogin", authApi.AuthHandler)

	}
}
