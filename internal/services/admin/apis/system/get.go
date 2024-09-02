package system

import (
	"context"
	"go-file-server/internal/common/core"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/config"
	"go-file-server/pkgs/utils/concurrentpool"
	"go-file-server/pkgs/utils/str"
	"go-file-server/pkgs/utils/timex"
	"go-file-server/pkgs/zlog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const systemKey string = "systemKey"

func (api *SystemApi) GetInfo(c *gin.Context) {
	ticker := timex.NewImmediateTicker(time.Second * 3)
	defer ticker.Stop()
	core.AutoStream(c, func(m core.MsgChan) {
		defer m.Close()
		for {
			select {
			case <-ticker.C:
				data, err := api.getData(c.Request.Context())
				if err != nil {
					zlog.SugLog.Error(err)
					m.Send("error", "内部服务异常")
					c.Writer.Flush()
					return
				}
				m.Send("message", data)
				c.Writer.Flush()
			case <-c.Request.Context().Done():
				return
			}
		}
	})

}

func (api *SystemApi) getData(ctx context.Context) (any, error) {
	api.mutex.RLock()
	s, err := api.cache.Get(systemKey)
	api.mutex.RUnlock()
	if err == nil || !cache.IsKeyNotFoundError(err) {
		return s, err
	}

	api.mutex.Lock()
	defer api.mutex.Unlock()
	s, err = api.cache.Get(systemKey)
	if err == nil || !cache.IsKeyNotFoundError(err) {
		return s, err
	}
	data, err := collectSystemDetails(ctx)
	if err != nil {
		return nil, err
	}
	strData, err := str.ConvertToString(data)
	if err != nil {
		return nil, err
	}
	err = api.cache.Set(systemKey, strData, time.Second*5)
	return data, err
}

type result struct {
	data any
	err  error
}

func collectSystemDetails(ctx context.Context) (*SystemDetails, error) {
	results := make(chan result, 4)
	pool, err := concurrentpool.NewAntsPool()
	defer pool.Release()
	if err != nil {
		return nil, err
	}
	path := config.ApplicationCfg.Basedir
	pool.Submit(func() { data, err := getOSInfo(ctx); results <- result{data, err} })
	pool.Submit(func() { data, err := getCPUInfo(ctx); results <- result{data, err} })
	pool.Submit(func() { data, err := getMemoryInfo(ctx); results <- result{data, err} })
	pool.Submit(func() { data, err := getDiskInfo(ctx, path); results <- result{data, err} })
	pool.Wait()
	close(results)

	details := &SystemDetails{}
	for {
		select {
		case result, ok := <-results:
			if !ok {
				return details, nil
			}
			if result.err != nil {
				zlog.SugLog.Error(err)
				continue
			}
			switch res := result.data.(type) {
			case *SystemDetails:
				details.OS = res.OS
				details.Uptime = res.Uptime
			case *CPUDetails:
				details.CPUInfo = *res
			case *MemoryDetails:
				details.MemoryInfo = *res
			case *DiskDetails:
				details.DiskInfo = *res
			default:
				zlog.SugLog.Warn("unknow result type: ", res)
			}

		case <-ctx.Done():
			zlog.SugLog.Error(ctx.Err())
			return details, nil
		}
	}
}

func getOSInfo(ctx context.Context) (*SystemDetails, error) {

	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &SystemDetails{
		OS:     info.OS,
		Uptime: info.Uptime,
	}, nil
}

func getCPUInfo(ctx context.Context) (*CPUDetails, error) {
	infos, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, err
	}
	if len(infos) > 0 {
		return &CPUDetails{
			UsagePercent: infos[0],
		}, nil
	}
	return &CPUDetails{}, nil
}

func getMemoryInfo(ctx context.Context) (*MemoryDetails, error) {
	memStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	free := memStat.Free + memStat.Buffers + memStat.Cached
	used := memStat.Total - free
	return &MemoryDetails{
		Total: memStat.Total,
		Used:  used,
		Free:  free,
	}, nil
}

func getDiskInfo(ctx context.Context, path string) (*DiskDetails, error) {
	if path == "" {
		path = "./"
	}
	diskStat, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return nil, err
	}
	return &DiskDetails{
		TotalSize:    diskStat.Total,
		UsedSize:     diskStat.Used,
		FreeSize:     diskStat.Free,
		UsagePercent: diskStat.UsedPercent,
	}, nil
}
