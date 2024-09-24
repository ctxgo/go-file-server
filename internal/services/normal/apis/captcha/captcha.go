package captcha

import (
	"go-file-server/internal/common/core"
	"go-file-server/pkgs/utils/captcha"

	"github.com/gin-gonic/gin"
)

type Captcha struct{}

func NewCaptchaAPI() *Captcha {
	return &Captcha{}
}

type CaptchaRep struct {
	Data   string `json:"data"`
	Id     string `json:"id"`
	Answer string
}

// GenerateCaptchaHandler 获取验证码
// @Summary 获取验证码
// @Description 获取验证码
// @Tags 登陆
// @Success 200 {object} response.Response{data=string,id=string,msg=string} "{"code": 200, "data": [...]}"
// @Router /api/v1/captcha [get]
func (e Captcha) GenerateCaptchaHandler(c *gin.Context) {

	id, b64s, Answer, err := captcha.DriverDigitFunc()
	if err != nil {
		c.Error(err)
		return
	}

	core.OKRep(
		CaptchaRep{
			Id:     id,
			Data:   b64s,
			Answer: Answer,
		},
	).SendGin(c)

}
