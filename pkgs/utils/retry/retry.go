package retry

import (
	"context"
	"fmt"
	"time"
)

// 定义错误检查函数类型
type retryableError func(error) bool

// 重试选项结构体
type RetryOptions struct {
	Attempts         int            // 最大尝试次数
	Delay            time.Duration  // 尝试之间的延迟
	IsRetryableError retryableError // 检查错误是否可重试的函数
}

// 设置重试次数的选项
func WithAttempts(n int) func(*RetryOptions) {
	return func(opts *RetryOptions) {
		opts.Attempts = n
	}
}

// 设置重试间隔的选项
func WithDelay(d time.Duration) func(*RetryOptions) {
	return func(opts *RetryOptions) {
		opts.Delay = d
	}
}

// 设置错误检查函数的选项
func WithRetryableErrorCheck(check retryableError) func(*RetryOptions) {
	return func(opts *RetryOptions) {
		opts.IsRetryableError = check
	}
}

func Retry(fn func() error, opts ...func(*RetryOptions)) error {
	return RetryWithCtx(context.Background(), fn, opts...)
}

// Retry 重试器，接收一个函数和多个配置选项
func RetryWithCtx(ctx context.Context, fn func() error, opts ...func(*RetryOptions)) error {
	options := RetryOptions{
		Attempts:         5,                                    // 默认重试次数
		Delay:            2 * time.Second,                      // 默认延迟
		IsRetryableError: func(err error) bool { return true }, // 默认始终重试
	}

	// 应用传入的选项
	for _, opt := range opts {
		opt(&options)
	}

	// 检查上下文是否已取消
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	var err error
	attempt := 0

	for attempt < options.Attempts || options.Attempts < 1 {
		attempt++
		err = fn()
		if err == nil {
			return nil
		}

		fmt.Printf("Attempt %d failed with error: %v\n", attempt, err) // 打印错误信息
		if !options.IsRetryableError(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(options.Delay):
		}
	}
	return err
}
