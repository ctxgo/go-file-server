package init

import (
	"go-file-server/pkgs/zlog"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func InitGin() *gin.Engine {
	//设置运行模式
	if !zlog.Log.Core().Enabled(zap.DebugLevel) {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化引擎
	r := gin.New()
	//r.RedirectTrailingSlash = false
	initGinValidator()
	return r
}

func initGinValidator() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(
		func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		},
	)
}
