package routers

import (
	"go-file-server/internal/services/normal/apis/captcha"

	"github.com/gin-gonic/gin"
)

func RegisterCaptchaRoutes(r gin.IRouter, captchaApi *captcha.Captcha) {
	{
		r.GET("captcha", captchaApi.GenerateCaptchaHandler)
	}
}
