package middlewares

import (
	"go-file-server/internal/common/core"
	"go-file-server/internal/common/global"
	"go-file-server/internal/common/repository"
	"go-file-server/internal/services/admin/models"
	"go-file-server/pkgs/zlog"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Glog struct {
	repo *repository.OperaLogRepository
}

func NewGlog(repo *repository.OperaLogRepository) Glog {
	return Glog{repo: repo}
}

func (g Glog) GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf := global.BufferPool.AcquireBuffer()
		defer global.BufferPool.ReleaseBuffer(buf)
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			tee := io.TeeReader(c.Request.Body, buf)
			c.Request.Body = io.NopCloser(tee)
		case http.MethodOptions:
			return
		}
		var data models.SysOperaLog

		start := time.Now()

		data.OperUrl = c.Request.URL.Path
		data.RequestMethod = c.Request.Method
		data.OperIp = c.ClientIP()
		data.UserAgent = c.Request.UserAgent()
		data.OperTime = start
		c.Next()
		claims := core.ExtractClaims(c)
		data.CreateBy = claims.UserId
		data.OperName = claims.Username
		data.LatencyTime = time.Since(start).String()
		data.Status = "1"
		if len(c.Errors) > 0 {
			data.Status = "2"
		}
		go func() {
			ptintLog(c, &data)
			g.intoDb(&data)
		}()
	}
}

func ptintLog(c *gin.Context, data *models.SysOperaLog) {

	logFields := []zap.Field{
		zap.String("method", data.RequestMethod),
		zap.String("path", data.OperUrl),
		zap.String("ip", data.OperIp),
		zap.String("user-agent", data.UserAgent),
		zap.String("cost", data.LatencyTime),
		zap.String("OperParam", data.OperParam),
	}
	zlog.Log.Info(strconv.Itoa(c.Writer.Status()), logFields...)

}

func (g Glog) intoDb(data *models.SysOperaLog) {
	g.repo.Create(data)

}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zlog.SugLog.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}
				zlog.SugLog.Error("[Recovery from panic]",
					zap.Any("error", err),
					zap.String("request", string(httpRequest)),
					zap.String("stack", string(debug.Stack())), //打印详细错误
				)
				//if stack {
				//	lg.SugLog.Error("[Recovery from panic]",
				//		zap.Any("error", err),
				//		zap.String("request", string(httpRequest)),
				//		zap.String("stack", string(debug.Stack())),
				//	)
				//} else {
				//	lg.SugLog.Error("[Recovery from panic]",
				//		zap.Any("error", err),
				//		zap.String("request", string(httpRequest)),
				//	)
				//}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
