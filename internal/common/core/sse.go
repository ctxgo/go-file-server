package core

import (
	"fmt"
	"go-file-server/internal/common/types"
	"go-file-server/pkgs/zlog"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SseErr struct {
	types.HttpRep
	rawErr error
}

func NewSseErr(e error) *SseErr {
	r := ErrRep()
	return &SseErr{
		HttpRep: r,
		rawErr:  e,
	}
}

func (e *SseErr) SetMsg(s string) *SseErr {
	e.HttpRep.SetMsg(s)
	return e
}

func (e *SseErr) GetRawErr() error {
	return e.rawErr
}

func (e *SseErr) Error() string {
	return e.HttpRep.GetMsg()
}

func OnceStream(c *gin.Context, k string, v any) {
	SetSseHeader(c)
	c.SSEvent(k, v)
	c.Writer.Flush()
}

func Stream(c *gin.Context, f func(MsgChan)) {
	stream(c, f, false)
}

// 每次发送数据后自动刷新
func AutoStream(c *gin.Context, f func(MsgChan)) {
	stream(c, f, true)
}

type Data struct {
	K string
	V any
}

type MsgChan chan Data

func (c MsgChan) Send(k string, v any) {
	c <- Data{K: k, V: v}
}

func (c MsgChan) Close() {
	close(c)
}

func stream(c *gin.Context, f func(MsgChan), autoFlush bool) {

	SetSseHeader(c)
	var logMsg string
	if zlog.Log.Level().Enabled(zap.DebugLevel) {
		logMsg = fmt.Sprintf("url: %s, ip: %s", c.Request.URL.Path, c.ClientIP())
	}

	msgChan := make(MsgChan)
	go f(msgChan)
	for data := range msgChan {
		if data.K != "" && data.V != nil {
			c.SSEvent(data.K, data.V)
			if autoFlush {
				c.Writer.Flush()
			}
			zlog.SugLog.Debugf("send, %s, data: %v", logMsg, data.V)
		}
	}
	zlog.SugLog.Debug("close, " + logMsg)
}

func SetSseHeader(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// 禁用nginx缓存
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}
