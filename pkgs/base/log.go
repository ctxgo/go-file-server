package base

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
)

// ZapLoggerAdapter adapts the zap.Logger to gorm.logger.Interface
type ZapLoggerAdapter struct {
	logger *zap.SugaredLogger
}

func NewZapLoggerAdapter(zapLogger *zap.SugaredLogger) logger.Interface {
	return &ZapLoggerAdapter{
		logger: zapLogger,
	}
}

func (z *ZapLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	// 返回一个新的实例，以支持在特定情境下修改日志级别（如果需要）
	newLogger := z.logger.WithOptions(zap.IncreaseLevel(zapcore.Level(level)))
	return &ZapLoggerAdapter{logger: newLogger}
}

func (z *ZapLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Infof(msg, data...)
}

func (z *ZapLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Warnf(msg, data...)
}

func (z *ZapLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	z.logger.Errorf(msg, data...)
}

func (z *ZapLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin).Milliseconds()
	sql, rows := fc()
	if err != nil {
		z.logger.Errorf("SQL error, sql: %v, rows: %d, elapsed: %d, err: %v", sql, rows, elapsed, err)
	} else {
		z.logger.Debugf("SQL Trace, sql: %v, rows: %d, elapsed: %d", sql, rows, elapsed)
	}
}
