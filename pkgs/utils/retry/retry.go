package retry

import (
	"context"
	"fmt"
	"time"
)

// 定义错误检查函数类型
type retryCondition func(error) bool

type ILogger interface {
	Warnf(template string, args ...interface{})
}

type Logger struct{}

func (l *Logger) Warnf(template string, args ...interface{}) {
	fmt.Printf(template, args...)
}

// 重试选项结构体
type RetryOptions struct {
	logger         ILogger
	attempts       int            // 最大尝试次数
	delay          time.Duration  // 尝试之间的延迟
	retryCondition retryCondition // 检查错误是否可重试的函数
}

// 设置重试次数的选项
func WithAttempts(n int) func(*RetryOptions) {
	return func(opts *RetryOptions) {
		opts.attempts = n
	}
}

// 设置logger
func WithLogger(l ILogger) func(*RetryOptions) {
	return func(opts *RetryOptions) {
		opts.logger = l
	}
}

// 设置重试间隔的选项
func WithDelay(d time.Duration) func(*RetryOptions) {
	return func(opts *RetryOptions) {
		opts.delay = d
	}
}

// 设置错误检查函数的选项
func WithRetryableCondition(f retryCondition) func(*RetryOptions) {
	return func(opts *RetryOptions) {
		opts.retryCondition = f
	}
}

func Retry(fn func() error, opts ...func(*RetryOptions)) error {
	return RetryWithCtx(context.Background(), fn, opts...)
}

// Retry 重试器，接收一个函数和多个配置选项
func RetryWithCtx(ctx context.Context, fn func() error, opts ...func(*RetryOptions)) error {
	options := RetryOptions{
		attempts:       5,                                    // 默认重试次数
		delay:          3 * time.Second,                      // 默认延迟
		retryCondition: func(err error) bool { return true }, // 默认始终重试
	}

	// 应用传入的选项
	for _, opt := range opts {
		opt(&options)
	}
	if options.logger == nil {
		options.logger = &Logger{}
	}
	// 检查上下文是否已取消
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	var err error
	attempt := 0

	for attempt < options.attempts || options.attempts < 1 {
		attempt++
		err = fn()
		if err == nil {
			return nil
		}

		options.logger.Warnf("Attempt %d failed with error: %+v\n", attempt, err) // 打印错误信息
		if !options.retryCondition(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(options.delay):
		}
	}
	return err
}
