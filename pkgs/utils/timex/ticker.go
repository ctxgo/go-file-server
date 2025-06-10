package timex

import (
	"context"
	"time"
)

// ImmediateTicker 在 time.Ticker 的基础上增加立即执行的功能
type ImmediateTicker struct {
	t      *time.Ticker
	C      chan time.Time
	cancel context.CancelFunc
}

// NewImmediateTicker 创建并返回一个立即触发的定时器
func NewImmediateTicker(interval time.Duration) *ImmediateTicker {
	ctx, cancel := context.WithCancel(context.Background())
	immediateTicker := &ImmediateTicker{
		t:      time.NewTicker(interval),
		C:      make(chan time.Time, 1),
		cancel: cancel,
	}
	go immediateTicker.start(ctx)
	return immediateTicker
}

func (it *ImmediateTicker) start(ctx context.Context) {
	defer it.t.Stop()
	defer close(it.C)
	it.C <- time.Now()
	for {
		select {
		case it.C <- <-it.t.C:
		case <-ctx.Done():
			return
		}
	}
}

// Stop 停止定时器
func (it *ImmediateTicker) Stop() {
	it.cancel()
}
