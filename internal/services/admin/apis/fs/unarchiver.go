package fs

import (
	"context"
	"go-file-server/internal/common/core"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/pkgs/utils/timex"
	"go-file-server/pkgs/zlog"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mholt/archiver/v4"
	"github.com/pkg/errors"
)

// 解压标识
const (
	// message
	unarchiveMessage = "message"
	// done
	unarchiveDone = "done"
	// error
	unarchiveError = "error"
)

// 解压
func (api *FsApi) Unarchive(c *gin.Context) {
	var req utils.UriPath
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.Error(err)
		return
	}
	realPath, err := utils.GetRealPath(req.Path)
	if err != nil {
		core.OnceStream(c, "error", err.Error())
		return
	}
	//如果ok, 说明已经有进程处理解压，直接订阅日志返回给客户端
	publisher, ok := api.publisherManager.Get(realPath)
	if ok {
		handleMsg(c, publisher)
		return
	}
	//如果 ok，说明已经被创建了，只需要订阅日志
	publisher, ok = api.publisherManager.GetOrSet(realPath, utils.NewPublisher[utils.Message]())
	if ok {
		handleMsg(c, publisher)
		return
	}
	defer api.publisherManager.Del(realPath)
	err = api.unarchive(c, realPath, publisher)
	if err != nil {
		c.Error(err)
		return
	}
}

func (api *FsApi) unarchive(c *gin.Context, realPath string, publisher *utils.Publisher[utils.Message]) (err error) {
	f, err := os.Open(realPath)
	if err != nil {
		if os.IsNotExist(err) {
			return core.NewSseErr(errors.WithStack(err)).
				SetMsg("待解压的文件已经不存在或移动到其他位置，请刷新界面")
		}
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = errors.WithStack(cerr)
		}
	}()
	extractor, err := parseExtractor(realPath, f)
	if err != nil {
		return errors.WithStack(err)
	}

	go api.execExtractor(c, extractor, f, realPath, publisher)
	handleMsg(c, publisher)

	return nil
}

func (api *FsApi) execExtractor(c *gin.Context, extractor archiver.Extractor,
	sourceArchive io.Reader, path string, publisher *utils.Publisher[utils.Message]) {
	defer publisher.Close()
	desName, desPath, err := createDesDir(path)
	if err != nil {
		publisher.Publish(utils.NewMessage("error", err.Error()))
		return
	}
	err = extractor.Extract(c.Request.Context(), sourceArchive, nil,
		func(ctx context.Context, f archiver.File) error {
			msg := filepath.Join(desName, f.NameInArchive)
			publisher.Publish(utils.NewMessage(unarchiveMessage, msg))
			return utils.HandleFile(ctx, f, desPath)
		},
	)
	if err != nil {
		zlog.SugLog.Error(err)
		if err == context.Canceled {
			publisher.Publish(utils.NewMessage(unarchiveError, "解压被取消"))
			return
		}
		publisher.Publish(utils.NewMessage(unarchiveError, "解压失败"))
		return
	}
	publisher.Publish(utils.NewMessage(unarchiveMessage, "更新索引..."))
	err = api.fsRepo.AddResource(desPath)
	if err != nil {
		zlog.SugLog.Error(err)
		publisher.Publish(utils.NewMessage(unarchiveError, "更新索引失败"))
		return
	}
	publisher.Publish(utils.NewMessage(unarchiveDone, "解压完成"))

}

func handleMsg(c *gin.Context, publisher *utils.Publisher[utils.Message]) {

	subscriber := publisher.CreateSubscriber()
	defer subscriber.Close()

	ticker := timex.NewImmediateTicker(time.Millisecond * 500)
	defer ticker.Stop()
	core.SetSseHeader(c)
	var currentMessage string
	defer c.Writer.Flush()
	for {
		select {
		case message := <-subscriber.Messages():
			currentMessage = message.K
			c.SSEvent(message.K, message.V)
		case <-ticker.C:
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		case <-publisher.Done():
			lastMessage := publisher.LastMessage()

			if !(lastMessage.K == unarchiveDone ||
				lastMessage.K == unarchiveError) {
				c.SSEvent(unarchiveError, "解压异常")
				return
			}
			if currentMessage != lastMessage.K {
				c.SSEvent(lastMessage.K, lastMessage.V)
			}
			return
		}

	}

}

func createDesDir(path string) (string, string, error) {
	baseName := filepath.Base(path)
	baseNames := strings.Split(baseName, ".")
	if len(baseNames) < 1 {
		return "", "", errors.New("不支持的压缩格式")

	}
	des := filepath.Join(filepath.Dir(path), baseNames[0])
	err := os.Mkdir(des, 0755)
	if err != nil {
		if os.IsExist(err) {
			return "", "", errors.New("解压失败,当前路径中已经存在: " + baseNames[0])
		}
		zlog.SugLog.Error(err)
		return "", "", errors.New("内部错误")
	}
	return baseNames[0], des, nil
}

func parseExtractor(filename string, stream io.Reader) (archiver.Extractor, error) {
	format, _, err := archiver.Identify(filename, stream)
	if err != nil {
		return nil, core.NewSseErr(err).SetMsg("识别格式失败")
	}

	extractor, ok := format.(archiver.Extractor)
	if !ok {
		return nil, core.NewSseErr(err).SetMsg("不支持该格式解压")
	}
	return extractor, nil
}
