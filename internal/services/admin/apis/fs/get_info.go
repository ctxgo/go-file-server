package fs

import (
	"go-file-server/internal/common/core"
	"go-file-server/pkgs/utils/timex"
	"go-file-server/pkgs/zlog"
	"time"

	"github.com/gin-gonic/gin"
)

const FsKey string = "fsCount"

type InfoRep struct {
	Count     uint64 `json:"count"`
	FileCount uint64 `json:"file_count"`
	DirCount  uint64 `json:"dir_count"`
}

func (api *FsApi) GetInfo(c *gin.Context) {

	ticker := timex.NewImmediateTicker(time.Second * 3)
	defer ticker.Stop()
	core.AutoStream(c, func(m core.MsgChan) {
		defer m.Close()
		for {
			select {
			case <-ticker.C:
				val, err := api.cache.Get(FsKey)
				if err != nil {
					zlog.SugLog.Error(err)
					m.Send("error", "内部服务异常")
					c.Writer.Flush()
					return
				}
				m.Send("message", val)
				c.Writer.Flush()
			case <-c.Request.Context().Done():
				return
			}
		}
	})

}
